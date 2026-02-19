package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aiox-platform/aiox/internal/database"
	mw "github.com/aiox-platform/aiox/internal/middleware"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

// HandlerSet holds handler functions injected from main.go to avoid import cycles.
type HandlerSet struct {
	// Auth handlers
	Register http.HandlerFunc
	Login    http.HandlerFunc
	Refresh  http.HandlerFunc
	Logout   http.HandlerFunc

	// Agent handlers
	CreateAgent          http.HandlerFunc
	ListAgents           http.HandlerFunc
	GetAgent             http.HandlerFunc
	UpdateAgent          http.HandlerFunc
	DeleteAgent          http.HandlerFunc
	OwnershipMiddleware  func(http.Handler) http.Handler

	// Memory handlers (Phase 4)
	ListMemories      http.HandlerFunc
	CreateMemory      http.HandlerFunc
	SearchMemories    http.HandlerFunc
	DeleteMemory      http.HandlerFunc
	DeleteAllMemories http.HandlerFunc

	// Governance handlers (Phase 5)
	GetUserQuota       http.HandlerFunc
	ListAuditLogs      http.HandlerFunc
	ListAgentAuditLogs http.HandlerFunc

	// Auth middleware
	AuthMiddleware func(http.Handler) http.Handler

	// Worker pool health (Phase 3)
	WorkerPoolHealthy func() bool
}

func NewRouter(pool *pgxpool.Pool, natsClient *inats.Client, h HandlerSet) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(mw.RequestID)
	r.Use(mw.Logging)
	r.Use(mw.Recovery)
	r.Use(cors.Handler(mw.CORS()))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]string{
			"status":   "healthy",
			"database": "healthy",
			"nats":     "healthy",
			"workers":  "healthy",
		}

		status := http.StatusOK

		if err := database.HealthCheck(r.Context(), pool); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
			status = http.StatusServiceUnavailable
		}

		if natsClient != nil && !natsClient.Healthy() {
			health["nats"] = "unhealthy"
			health["status"] = "degraded"
			status = http.StatusServiceUnavailable
		} else if natsClient == nil {
			health["nats"] = "not configured"
		}

		if h.WorkerPoolHealthy != nil {
			if !h.WorkerPoolHealthy() {
				health["workers"] = "no workers connected"
				health["status"] = "degraded"
			}
		} else {
			health["workers"] = "not configured"
		}

		JSON(w, status, health)
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.Refresh)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Post("/logout", h.Logout)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			// Agent routes
			r.Route("/agents", func(r chi.Router) {
				r.Post("/", h.CreateAgent)
				r.Get("/", h.ListAgents)

				r.Route("/{agentID}", func(r chi.Router) {
					r.Use(h.OwnershipMiddleware)
					r.Get("/", h.GetAgent)
					r.Put("/", h.UpdateAgent)
					r.Delete("/", h.DeleteAgent)

					// Memory routes (Phase 4)
					r.Route("/memories", func(r chi.Router) {
						r.Get("/", h.ListMemories)
						r.Post("/", h.CreateMemory)
						r.Post("/search", h.SearchMemories)
						r.Delete("/", h.DeleteAllMemories)
						r.Delete("/{memoryID}", h.DeleteMemory)
					})

					// Agent audit logs (Phase 5)
					r.Get("/audit", h.ListAgentAuditLogs)
				})
			})

			// Governance routes (Phase 5)
			r.Route("/governance", func(r chi.Router) {
				r.Get("/quota", h.GetUserQuota)
				r.Get("/audit", h.ListAuditLogs)
			})
		})
	})

	return r
}
