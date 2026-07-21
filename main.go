package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/leventkok/mlc-llm-monitoring/internal/auth"
	"github.com/leventkok/mlc-llm-monitoring/internal/database"
	"github.com/leventkok/mlc-llm-monitoring/internal/handlers"
	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
)

func main() {
	_ = godotenv.Load()

	auth.InitJWT()
	middleware.EnsureCORS()

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()
	fmt.Println("Connected to PostgreSQL")

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("Database schema ready")

	store := storage.NewPostgresStore(pool)
	configStore := storage.NewConfigStore()

	authHandler := handlers.NewAuthHandler(store)
	configHandler := handlers.NewConfigHandler(configStore)
	reviewHandler := handlers.NewReviewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health(pool))
	mux.HandleFunc("/ready", handlers.Health(pool))

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configHandler.Get(w, r)
		case http.MethodPut:
			middleware.RequireAuth(configHandler.Update)(w, r)
		default:
			writeMethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/me", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authHandler.Me(w, r)
		case http.MethodPatch:
			authHandler.UpdateMe(w, r)
		case http.MethodDelete:
			authHandler.DeleteMe(w, r)
		default:
			writeMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/auth/logout", middleware.RequireAuth(authHandler.Logout))
	mux.HandleFunc("/auth/refresh", middleware.RequireAuth(authHandler.Refresh))
	mux.HandleFunc("/auth/validate", middleware.RequireAuth(authHandler.Validate))
	mux.HandleFunc("/auth/change-password", middleware.RequireAuth(authHandler.ChangePassword))

	mux.HandleFunc("GET /reviews/{id}", middleware.RequireAuth(reviewHandler.GetReview))

	mux.HandleFunc("/reviews", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			reviewHandler.ListReviews(w, r)
		case http.MethodPost:
			reviewHandler.CreateReview(w, r)
		default:
			writeMethodNotAllowed(w)
		}
	}))

	mux.HandleFunc("/decisions", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			reviewHandler.ListDecisions(w, r)
		case http.MethodPost:
			reviewHandler.SaveDecision(w, r)
		default:
			writeMethodNotAllowed(w)
		}
	}))

	mux.HandleFunc("/scores", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			reviewHandler.ListScores(w, r)
		case http.MethodPost:
			reviewHandler.CreateScore(w, r)
		default:
			writeMethodNotAllowed(w)
		}
	}))

	mux.HandleFunc("/metrics", middleware.RequireAuth(reviewHandler.GetMetrics))

	handler := middleware.Chain(
		mux,
		middleware.SecurityHeaders,
		middleware.MaxBodyBytes(1<<20),
		middleware.RequestTimeout,
		middleware.RateLimit(10, time.Minute, "/auth/login", "/auth/register"),
		middleware.CORS,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		fmt.Printf("app-review-monitoring API started on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	fmt.Println("server stopped")
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte(`{"error":"unsupported method"}`))
}
