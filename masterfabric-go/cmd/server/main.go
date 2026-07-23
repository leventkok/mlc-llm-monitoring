package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	iamUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/usecase"
	configUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/config/usecase"
	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
	memConfig "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/config/memory"
	configHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/config"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/health"
	iamHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/iam"
	llmHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/llm"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/router"
	infraMLC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/mlc"
	pgIam "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/iam"
	pgLlm "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/llm"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/config"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/database"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/logger"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/telemetry"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/version"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := config.Load()
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	validateJWTSecret(cfg.JWT.Secret)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.NewPostgresPool(ctx, *cfg)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer db.Close()
	log.Info("connected to postgres")

	if err := database.MigrateAppSchema(ctx, db); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Info("database schema ready")

	if err := health.PingDB(ctx, db); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	var metricsHandler http.Handler
		if cfg.Telemetry.Enabled {
		shutdownTelemetry, err := telemetry.Setup(ctx, cfg.Telemetry.ServiceName, version.Version)
		if err != nil {
			return fmt.Errorf("telemetry setup failed: %w", err)
		}
		defer func() {
			_ = shutdownTelemetry(context.Background())
		}()
		metricsHandler = promhttp.Handler()
		log.Info("prometheus metrics enabled", "path", "/metrics")
	}

	appJWT := infraAuth.NewAppJWTService(cfg.JWT.Secret)
	deps := buildDependencies(log, cfg, db, appJWT, metricsHandler)

	handler := router.New(deps)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Info("app-review-monitoring API started", "addr", addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case <-shutdown:
		log.Info("shutdown signal received")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		log.Info("server stopped")
	}

	return nil
}

func validateJWTSecret(secret string) {
	if os.Getenv("JWT_SECRET") == "" {
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("JWT_SECRET must be set in production")
		}
		log.Println("WARNING: JWT_SECRET not set, using dev-only fallback")
	}
	if os.Getenv("GO_ENV") == "production" && len(secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters in production")
	}
}

func buildDependencies(log *slog.Logger, cfg *config.Config, db *pgxpool.Pool, appJWT *infraAuth.AppJWTService, metricsHandler http.Handler) router.Dependencies {
	userRepo := pgIam.NewAppUserRepo(db)
	reviewRepo := pgLlm.NewReviewRepo(db)
	configRepo := memConfig.NewConfigRepo()

	registerUC := iamUC.NewRegisterUseCase(userRepo)
	loginUC := iamUC.NewLoginUseCase(userRepo, appJWT)
	getMeUC := iamUC.NewGetMeUseCase(userRepo)
	updateMeUC := iamUC.NewUpdateMeUseCase(userRepo)
	deleteMeUC := iamUC.NewDeleteMeUseCase(userRepo)
	refreshUC := iamUC.NewRefreshUseCase(appJWT)
	changePasswordUC := iamUC.NewChangePasswordUseCase(userRepo)

	createReviewUC := llmUC.NewCreateReviewUseCase(reviewRepo)
	getReviewUC := llmUC.NewGetReviewUseCase(reviewRepo)
	listReviewsUC := llmUC.NewListReviewsUseCase(reviewRepo)
	createDecisionUC := llmUC.NewCreateDecisionUseCase(reviewRepo)
	listDecisionsUC := llmUC.NewListDecisionsUseCase(reviewRepo)
	createScoreUC := llmUC.NewCreateScoreUseCase(reviewRepo)
	listScoresUC := llmUC.NewListScoresUseCase(reviewRepo)
	getMetricsUC := llmUC.NewGetMetricsUseCase(reviewRepo)

	var analyzeReviewUC *llmUC.AnalyzeReviewUseCase
	if cfg.MLC.Enabled {
		mlcClient := infraMLC.NewClient(cfg.MLC.BaseURL, cfg.MLC.Model, cfg.MLC.APIKey)
		analyzeReviewUC = llmUC.NewAnalyzeReviewUseCase(reviewRepo, mlcClient)
		log.Info("server-side mlc inference enabled", "base_url", cfg.MLC.BaseURL, "model", cfg.MLC.Model)
	}

	getConfigUC := configUC.NewGetConfigUseCase(configRepo)
	updateConfigUC := configUC.NewUpdateConfigUseCase(configRepo)

	return router.Dependencies{
		Logger:             log,
		DB:                 db,
		CORSAllowedOrigins: cfg.Server.CORSAllowedOrigins,
		MaxBodyBytes:       cfg.Server.MaxBodyBytes,
		AppJWT:             appJWT,
		MetricsHandler:     metricsHandler,
		IAMHandler: iamHandler.NewHandler(
			registerUC, loginUC, getMeUC, updateMeUC, deleteMeUC, refreshUC, changePasswordUC,
		),
		LLMHandler: llmHandler.NewHandler(
			createReviewUC, getReviewUC, listReviewsUC, analyzeReviewUC,
			createDecisionUC, listDecisionsUC,
			createScoreUC, listScoresUC, getMetricsUC,
		),
		ConfigHandler: configHandler.NewHandler(getConfigUC, updateConfigUC),
	}
}
