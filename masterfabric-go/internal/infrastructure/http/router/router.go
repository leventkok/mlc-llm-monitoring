package router

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	configHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/config"
	datasetHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/dataset"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/health"
	adminHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/admin"
	agentHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/agent"
	inviteHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/invite"
	orgHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/org"
	orgadminHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/orgadmin"
	iamHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/iam"
	llmHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/llm"
	mcpHandler "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/http/handler/mcp"
	infraAuth "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/auth"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/middleware"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/response"
)

// Dependencies holds injected dependencies for the app router.
type Dependencies struct {
	Logger             *slog.Logger
	DB                 *pgxpool.Pool
	CORSAllowedOrigins []string
	MaxBodyBytes       int64
	MetricsEnabled     bool
	AppJWT             *infraAuth.AppJWTService

	IAMHandler    *iamHandler.Handler
	LLMHandler    *llmHandler.Handler
	ConfigHandler  *configHandler.Handler
	DatasetHandler *datasetHandler.Handler
	AdminHandler  *adminHandler.Handler
	OrgAdminHandler *orgadminHandler.Handler
	OrgHandler      *orgHandler.Handler
	InviteHandler *inviteHandler.Handler
	AgentHandler  *agentHandler.Handler
	MLCAPIKey     string
	MCPHandler    *mcpHandler.Handler
	UserRepo      repository.UserRepository
	MetricsHandler http.Handler
}

// New creates the root Chi router with legacy frontend API routes at root paths.
func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(deps.Logger))
	r.Use(middleware.Recoverer(deps.Logger))
	if deps.MaxBodyBytes > 0 {
		r.Use(middleware.MaxBodyBytes(deps.MaxBodyBytes))
	}
	r.Use(cors.Handler(middleware.CORSOptions(deps.CORSAllowedOrigins)))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RequestTimeoutLegacy)
	if deps.MetricsEnabled {
		r.Use(middleware.Prometheus)
	}
	r.Use(middleware.LegacyRateLimit(10, time.Minute, "/auth/login", "/auth/register"))

	legacyHealth := health.LegacyHealth(deps.DB)
	r.Get("/health", legacyHealth)
	r.Get("/ready", legacyHealth)

	if deps.MetricsHandler != nil {
		r.Handle("/metrics", deps.MetricsHandler)
	}

	if deps.ConfigHandler != nil {
		r.Get("/config", deps.ConfigHandler.Get)
		r.With(middleware.LegacyAuth(deps.AppJWT)).Put("/config", deps.ConfigHandler.Update)
	}

	if deps.IAMHandler != nil {
		r.Post("/auth/register", deps.IAMHandler.Register)
		r.Post("/auth/login", deps.IAMHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Get("/auth/me", deps.IAMHandler.Me)
			r.Patch("/auth/me", deps.IAMHandler.UpdateMe)
			r.Delete("/auth/me", deps.IAMHandler.DeleteMe)
			r.Post("/auth/logout", deps.IAMHandler.Logout)
			r.Post("/auth/refresh", deps.IAMHandler.Refresh)
			r.Get("/auth/validate", deps.IAMHandler.Validate)
			r.Post("/auth/change-password", deps.IAMHandler.ChangePassword)
		})
	}

	if deps.LLMHandler != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Get("/reviews/{id}", deps.LLMHandler.GetReview)
			r.Post("/reviews/{id}/analyze", deps.LLMHandler.AnalyzeReview)
			r.Get("/reviews", deps.LLMHandler.ListReviews)
			r.Post("/reviews", deps.LLMHandler.CreateReview)
			r.Get("/decisions", deps.LLMHandler.ListDecisions)
			r.Post("/decisions", deps.LLMHandler.SaveDecision)
			r.Get("/scores", deps.LLMHandler.ListScores)
			r.Post("/scores", deps.LLMHandler.CreateScore)
			r.Get("/stats", deps.LLMHandler.GetMetrics)
		})
	}

	if deps.DatasetHandler != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Use(middleware.LegacyRequireAdmin(deps.UserRepo))
			r.Get("/dataset/reviews", deps.DatasetHandler.ListReviews)
			r.Get("/dataset/reviews.csv", deps.DatasetHandler.ExportCSV)
			r.Post("/dataset/batch-analyze", deps.DatasetHandler.BatchAnalyze)
		})
	}

	if deps.InviteHandler != nil {
		r.Get("/invites/{token}", deps.InviteHandler.Preview)
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Post("/invites/{token}/accept", deps.InviteHandler.Accept)
		})
	}

	if deps.OrgHandler != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Get("/organization/members", deps.OrgHandler.ListMembers)
			r.Delete("/organization/members/{userId}", deps.OrgHandler.RemoveMember)
			r.Get("/organization/invites", deps.OrgHandler.ListInvites)
			r.Post("/organization/invites", deps.OrgHandler.CreateInvite)
		})
	}

	if deps.MCPHandler != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Post("/mcp", deps.MCPHandler.Handle)
		})
	}

	if deps.AdminHandler != nil && deps.UserRepo != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Use(middleware.LegacyRequireAdmin(deps.UserRepo))
			r.Get("/admin/llm-config", deps.AdminHandler.GetLLMConfig)
			r.Put("/admin/llm-config", deps.AdminHandler.UpdateLLMConfig)
			r.Get("/admin/analyze-logs", deps.AdminHandler.AnalyzeLogs)
			r.Get("/admin/model-profiles", deps.AdminHandler.ListModelProfiles)
			r.Post("/admin/switch-model", deps.AdminHandler.SwitchModel)
			r.Get("/admin/model-switch/status", deps.AdminHandler.ModelSwitchStatus)
		})
	}

	if deps.OrgAdminHandler != nil && deps.UserRepo != nil {
		r.Group(func(r chi.Router) {
			r.Use(middleware.LegacyAuth(deps.AppJWT))
			r.Use(middleware.LegacyRequireAdmin(deps.UserRepo))
			r.Get("/admin/organizations", deps.OrgAdminHandler.ListOrganizations)
			r.Post("/admin/organizations", deps.OrgAdminHandler.CreateOrganization)
			r.Delete("/admin/organizations/{orgId}", deps.OrgAdminHandler.DeleteOrganization)
			r.Get("/admin/organizations/{orgId}/invites", deps.OrgAdminHandler.ListInvites)
			r.Post("/admin/organizations/{orgId}/invites", deps.OrgAdminHandler.CreateInvite)
		})
	}

	if deps.AgentHandler != nil && strings.TrimSpace(deps.MLCAPIKey) != "" {
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireMLCAPIKey(deps.MLCAPIKey))
			r.Get("/agent/model-switch/next", deps.AgentHandler.ClaimNext)
			r.Post("/agent/model-switch/{id}/report", deps.AgentHandler.Report)
		})
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.LegacyError(w, http.StatusNotFound, "not found")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.LegacyError(w, http.StatusMethodNotAllowed, "unsupported method")
	})

	return r
}
