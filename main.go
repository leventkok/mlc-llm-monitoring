package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/leventkok/mlc-llm-monitoring/internal/database"
	"github.com/leventkok/mlc-llm-monitoring/internal/handlers"
	"github.com/leventkok/mlc-llm-monitoring/internal/middleware"
	"github.com/leventkok/mlc-llm-monitoring/internal/storage"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
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

	http.HandleFunc("/health", handlers.Health)

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configHandler.Get(w, r)
		case http.MethodPut:
			configHandler.Update(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"unsupported method"}`))
		}
	})

	
	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/login", authHandler.Login)
	http.HandleFunc("/auth/me", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authHandler.Me(w, r)
		case http.MethodPatch:
			authHandler.UpdateMe(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"unsupported method"}`))
		}
	}))
	http.HandleFunc("/auth/logout", authHandler.Logout)
	http.HandleFunc("/auth/refresh", middleware.RequireAuth(authHandler.Refresh))
	http.HandleFunc("/auth/validate", middleware.RequireAuth(authHandler.Validate))
	http.HandleFunc("/auth/change-password", middleware.RequireAuth(authHandler.ChangePassword))

	handler := middleware.CORS(http.DefaultServeMux)
	fmt.Println("Server started: http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}