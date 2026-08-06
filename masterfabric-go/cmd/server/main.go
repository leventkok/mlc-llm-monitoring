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
	datasetUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/dataset/usecase"
	configUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/config/usecase"
	configModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
	llmScope "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/scope"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
	memConfig "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/config/memory"
	adminHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/admin"
	configHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/config"
	datasetHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/dataset"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/health"
	iamHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/iam"
	llmHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/llm"
	agentHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/agent"
	mcpHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/mcp"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/router"
	infraDeepWiki "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/deepwiki"
	infraHFDataset "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/hfdataset"
	infraMLC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/mlc"
	orgUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/usecase"
	inviteHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/invite"
	orgHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/org"
	orgadminHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/orgadmin"
	pgOrg "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/org"
	pgConfig "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/config"
	pgIam "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/iam"
	pgLlm "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/llm"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/config"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/database"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/logger"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/metrics"
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
	if err := database.BackfillRoles(ctx, db); err != nil {
		return fmt.Errorf("role backfill failed: %w", err)
	}
	if err := database.BackfillMissingAutoScores(ctx, db); err != nil {
		return fmt.Errorf("score backfill failed: %w", err)
	}
	log.Info("database schema ready")

	if err := health.PingDB(ctx, db); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	var metricsHandler http.Handler
	metricsEnabled := cfg.Telemetry.Enabled
	if metricsEnabled {
		metrics.Register()
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
	deps := buildDependencies(log, cfg, db, appJWT, metricsHandler, metricsEnabled)

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

func buildDependencies(log *slog.Logger, cfg *config.Config, db *pgxpool.Pool, appJWT *infraAuth.AppJWTService, metricsHandler http.Handler, metricsEnabled bool) router.Dependencies {
	userRepo := pgIam.NewAppUserRepo(db)
	orgRepo := pgOrg.NewRepository(db)
	orgService := orgUC.NewService(orgRepo).WithUsers(userRepo)
	reviewScope := llmScope.NewResolver(orgService)
	reviewRepo := pgLlm.NewReviewRepo(db)
	configRepo := memConfig.NewConfigRepo()

	registerUC := iamUC.NewRegisterUseCase(userRepo)
	loginUC := iamUC.NewLoginUseCase(userRepo, appJWT)
	getMeUC := iamUC.NewGetMeUseCase(userRepo, orgService)
	updateMeUC := iamUC.NewUpdateMeUseCase(userRepo, orgService)
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
	var mlcClient *infraMLC.Client
	switchRepo := pgConfig.NewSwitchRepo(db)
	if cfg.MLC.Enabled {
		configRepo.SetRuntimeLLM(cfg.MLC.Model, os.Getenv("MLC_LORA_ADAPTER"))
		if latest, err := switchRepo.GetLatest(context.Background()); err == nil && latest != nil &&
			latest.Status == configModel.SwitchStatusCompleted {
			configRepo.SwitchLLM(latest.RequestModel, latest.LocalAdapter)
			log.Info("restored active model from last switch", "profile", latest.ProfileID, "model", latest.RequestModel)
		}
		mlcClient = infraMLC.NewClient(cfg.MLC.BaseURL, cfg.MLC.Model, cfg.MLC.APIKey, configRepo)
		analyzeReviewUC = llmUC.NewAnalyzeReviewUseCase(reviewRepo, mlcClient)
		log.Info("server-side mlc inference enabled", "base_url", cfg.MLC.BaseURL, "model", cfg.MLC.Model)
	}

	hfDatasetClient := infraHFDataset.NewClient(
		os.Getenv("HF_DATASET_ID"),
		os.Getenv("HF_TOKEN"),
		os.Getenv("HF_DATASETS_SERVER"),
		os.Getenv("DATASET_WORKER_URL"),
	)
	listDatasetUC := datasetUC.NewListDatasetReviewsUseCase(hfDatasetClient)
	exportDatasetUC := datasetUC.NewExportDatasetCSVUseCase(hfDatasetClient)
	var batchDatasetUC *datasetUC.BatchAnalyzeDatasetUseCase
	if mlcClient != nil {
		batchDatasetUC = datasetUC.NewBatchAnalyzeDatasetUseCase(hfDatasetClient, mlcClient)
	}

	deepWikiClient := infraDeepWiki.NewClient(os.Getenv("DEEPWIKI_MCP_URL"))
	mcpHTTPHandler := mcpHandler.NewHandler(reviewScope, analyzeReviewUC, deepWikiClient, batchDatasetUC)

	getConfigUC := configUC.NewGetConfigUseCase(configRepo)
	updateConfigUC := configUC.NewUpdateConfigUseCase(configRepo)
	getLLMConfigUC := configUC.NewGetLLMConfigUseCase(configRepo)
	updateLLMConfigUC := configUC.NewUpdateLLMConfigUseCase(configRepo)
	listProfilesUC := configUC.NewListModelProfilesUseCase(nil)
	switchModelUC := configUC.NewSwitchModelUseCase(listProfilesUC, configRepo, switchRepo)
	switchStatusUC := configUC.NewGetModelSwitchStatusUseCase(switchRepo)
	agentSwitchUC := configUC.NewAgentModelSwitchUseCase(switchRepo)

	return router.Dependencies{
		Logger:             log,
		DB:                 db,
		CORSAllowedOrigins: cfg.Server.CORSAllowedOrigins,
		MaxBodyBytes:       cfg.Server.MaxBodyBytes,
		MetricsEnabled:     metricsEnabled,
		AppJWT:             appJWT,
		MetricsHandler:     metricsHandler,
		UserRepo:           userRepo,
		IAMHandler: iamHandler.NewHandler(
			registerUC, loginUC, getMeUC, updateMeUC, deleteMeUC, refreshUC, changePasswordUC,
		),
		LLMHandler: llmHandler.NewHandler(
			reviewScope,
			createReviewUC, getReviewUC, listReviewsUC, analyzeReviewUC,
			createDecisionUC, listDecisionsUC,
			createScoreUC, listScoresUC, getMetricsUC,
		),
		ConfigHandler: configHandler.NewHandler(getConfigUC, updateConfigUC),
		DatasetHandler: datasetHandler.NewHandler(listDatasetUC, exportDatasetUC, batchDatasetUC),
		AdminHandler: adminHandler.NewHandler(
			getLLMConfigUC, updateLLMConfigUC, listProfilesUC, switchModelUC, switchStatusUC,
		),
		OrgAdminHandler: orgadminHandler.NewHandler(orgService),
		OrgHandler:      orgHandler.NewHandler(orgService),
		InviteHandler:   inviteHandler.NewHandler(orgService),
		AgentHandler:  agentHandler.NewHandler(agentSwitchUC),
		MLCAPIKey:     cfg.MLC.APIKey,
		MCPHandler:    mcpHTTPHandler,
	}
}
