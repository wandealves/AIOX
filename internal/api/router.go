package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	CreateAgent         http.HandlerFunc
	ListAgents          http.HandlerFunc
	GetAgent            http.HandlerFunc
	UpdateAgent         http.HandlerFunc
	DeleteAgent         http.HandlerFunc
	OwnershipMiddleware func(http.Handler) http.Handler

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

	// Tool handlers (Phase 11)
	CreateTool http.HandlerFunc
	ListTools  http.HandlerFunc
	GetTool    http.HandlerFunc
	UpdateTool http.HandlerFunc
	DeleteTool http.HandlerFunc

	// Chat handlers (Phase 7)
	SendMessage       http.HandlerFunc
	ListConversations http.HandlerFunc

	// WebSocket handler (Phase 7)
	WSUpgrade http.HandlerFunc

	// Organization handlers (Phase 9)
	CreateOrg          http.HandlerFunc
	ListOrgs           http.HandlerFunc
	GetOrg             http.HandlerFunc
	UpdateOrg          http.HandlerFunc
	DeleteOrg          http.HandlerFunc
	ListOrgMembers     http.HandlerFunc
	UpdateMemberRole   http.HandlerFunc
	RemoveOrgMember    http.HandlerFunc
	CreateOrgInvite    http.HandlerFunc
	ListOrgInvites     http.HandlerFunc
	DeleteOrgInvite    http.HandlerFunc
	AcceptInvite       http.HandlerFunc
	ListOrgAgents      http.HandlerFunc
	CreateOrgAgent     http.HandlerFunc
	OrgMemberMiddleware func(http.Handler) http.Handler
	OrgAdminMiddleware  func(http.Handler) http.Handler
	OrgOwnerMiddleware  func(http.Handler) http.Handler

	// Pipeline handlers (Phase 11B)
	CreatePipeline  http.HandlerFunc
	ListPipelines   http.HandlerFunc
	GetPipeline     http.HandlerFunc
	UpdatePipeline  http.HandlerFunc
	DeletePipeline  http.HandlerFunc
	ExecutePipeline http.HandlerFunc
	ListExecutions  http.HandlerFunc
	GetExecution    http.HandlerFunc

	// Scheduler handlers (Phase 11C)
	CreateSchedule http.HandlerFunc
	ListSchedules  http.HandlerFunc
	GetSchedule    http.HandlerFunc
	UpdateSchedule http.HandlerFunc
	DeleteSchedule http.HandlerFunc

	// Attachment handlers (Phase 11D)
	UploadAttachment   http.HandlerFunc
	ListAttachments    http.HandlerFunc
	GetAttachment      http.HandlerFunc
	DownloadAttachment http.HandlerFunc
	DeleteAttachment   http.HandlerFunc

	// Auth middleware
	AuthMiddleware func(http.Handler) http.Handler

	// Worker pool health (Phase 3)
	WorkerPoolHealthy func() bool
}

// RouterConfig holds configuration for the router.
type RouterConfig struct {
	CORSAllowedOrigins []string
	AuthRateLimiter    func(http.Handler) http.Handler
}

func NewRouter(pool *pgxpool.Pool, natsClient *inats.Client, cfg RouterConfig, h HandlerSet) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(mw.RequestID)
	r.Use(mw.SecurityHeaders)
	r.Use(mw.Logging)
	r.Use(mw.Recovery)
	r.Use(mw.Metrics)
	r.Use(cors.Handler(mw.CORS(cfg.CORSAllowedOrigins)))

	// Liveness probe — always 200, no dependency checks
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})

	// Readiness probe — checks DB, NATS, workers
	readinessHandler := func(w http.ResponseWriter, r *http.Request) {
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
	}

	r.Get("/health/ready", readinessHandler)
	r.Get("/health", readinessHandler)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public) — optionally rate-limited
		r.Route("/auth", func(r chi.Router) {
			if cfg.AuthRateLimiter != nil {
				r.Use(cfg.AuthRateLimiter)
			}
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.Refresh)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(h.AuthMiddleware)
				r.Post("/logout", h.Logout)
			})
		})

		// WebSocket upgrade — auth is handled internally via query param token
		if h.WSUpgrade != nil {
			r.Get("/ws", h.WSUpgrade)
		}

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

					// Chat routes (Phase 7)
					if h.SendMessage != nil {
						r.Post("/messages", h.SendMessage)
					}
					if h.ListConversations != nil {
						r.Get("/conversations", h.ListConversations)
					}

					// Tool routes (Phase 11)
					if h.CreateTool != nil {
						r.Route("/tools", func(r chi.Router) {
							r.Post("/", h.CreateTool)
							r.Get("/", h.ListTools)
							r.Get("/{toolID}", h.GetTool)
							r.Put("/{toolID}", h.UpdateTool)
							r.Delete("/{toolID}", h.DeleteTool)
						})
					}

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

			// Pipeline routes (Phase 11B)
			if h.CreatePipeline != nil {
				r.Route("/pipelines", func(r chi.Router) {
					r.Post("/", h.CreatePipeline)
					r.Get("/", h.ListPipelines)

					r.Route("/{pipelineID}", func(r chi.Router) {
						r.Get("/", h.GetPipeline)
						r.Put("/", h.UpdatePipeline)
						r.Delete("/", h.DeletePipeline)
						r.Post("/execute", h.ExecutePipeline)
						r.Get("/executions", h.ListExecutions)
						r.Get("/executions/{executionID}", h.GetExecution)
					})
				})
			}

			// Attachment routes (Phase 11D)
			if h.UploadAttachment != nil {
				r.Route("/attachments", func(r chi.Router) {
					r.Post("/", h.UploadAttachment)
					r.Get("/", h.ListAttachments)

					r.Route("/{attachmentID}", func(r chi.Router) {
						r.Get("/", h.GetAttachment)
						r.Get("/download", h.DownloadAttachment)
						r.Delete("/", h.DeleteAttachment)
					})
				})
			}

			// Scheduler routes (Phase 11C)
			if h.CreateSchedule != nil {
				r.Route("/schedules", func(r chi.Router) {
					r.Post("/", h.CreateSchedule)
					r.Get("/", h.ListSchedules)

					r.Route("/{scheduleID}", func(r chi.Router) {
						r.Get("/", h.GetSchedule)
						r.Put("/", h.UpdateSchedule)
						r.Delete("/", h.DeleteSchedule)
					})
				})
			}

			// Organization routes (Phase 9)
			if h.CreateOrg != nil {
				r.Post("/orgs", h.CreateOrg)
				r.Get("/orgs", h.ListOrgs)

				r.Route("/orgs/{orgID}", func(r chi.Router) {
					r.Use(h.OrgMemberMiddleware)
					r.Get("/", h.GetOrg)
					r.Get("/members", h.ListOrgMembers)
					r.Get("/agents", h.ListOrgAgents)

					// Admin+ routes
					r.Group(func(r chi.Router) {
						r.Use(h.OrgAdminMiddleware)
						r.Put("/", h.UpdateOrg)
						r.Put("/members/{userID}", h.UpdateMemberRole)
						r.Delete("/members/{userID}", h.RemoveOrgMember)
						r.Post("/invites", h.CreateOrgInvite)
						r.Get("/invites", h.ListOrgInvites)
						r.Delete("/invites/{inviteID}", h.DeleteOrgInvite)
						r.Post("/agents", h.CreateOrgAgent)
					})

					// Owner-only routes
					r.Group(func(r chi.Router) {
						r.Use(h.OrgOwnerMiddleware)
						r.Delete("/", h.DeleteOrg)
					})
				})

				// Invite acceptance (outside org routes — user just needs auth)
				r.Post("/invites/{token}/accept", h.AcceptInvite)
			}
		})
	})

	return r
}
