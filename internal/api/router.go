package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aiox-platform/aiox/internal/database"
	mw "github.com/aiox-platform/aiox/internal/middleware"
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

	// Auth middleware
	AuthMiddleware func(http.Handler) http.Handler
}

func NewRouter(pool *pgxpool.Pool, h HandlerSet) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(mw.RequestID)
	r.Use(mw.Logging)
	r.Use(mw.Recovery)
	r.Use(cors.Handler(mw.CORS()))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(r.Context(), pool); err != nil {
			JSONErrorMessage(w, http.StatusServiceUnavailable, "database unhealthy")
			return
		}
		JSON(w, http.StatusOK, map[string]string{"status": "healthy"})
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
				})
			})
		})
	})

	return r
}
