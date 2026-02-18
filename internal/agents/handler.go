package agents

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:      svc,
		validate: validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	agent, err := h.svc.Create(r.Context(), ownerID, &req)
	if err != nil {
		slog.Error("creating agent", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusCreated, agent)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	params := DefaultListParams()
	if p := r.URL.Query().Get("page"); p != "" {
		if page, err := strconv.Atoi(p); err == nil && page > 0 {
			params.Page = page
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if pageSize, err := strconv.Atoi(ps); err == nil && pageSize > 0 && pageSize <= 100 {
			params.PageSize = pageSize
		}
	}

	agents, totalCount, err := h.svc.ListByOwner(r.Context(), ownerID, params)
	if err != nil {
		slog.Error("listing agents", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONPaginated(w, http.StatusOK, agents, totalCount, params.Page, params.PageSize)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	agent := GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	api.JSON(w, http.StatusOK, agent)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	agent := GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	var req UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	updated, err := h.svc.Update(r.Context(), agent, &req)
	if err != nil {
		slog.Error("updating agent", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	agent := GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	if err := h.svc.Delete(r.Context(), agent.ID); err != nil {
		slog.Error("deleting agent", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "agent deleted successfully")
}

// OwnershipMiddleware verifies agent ownership before allowing access.
func (h *Handler) OwnershipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetUserClaims(r.Context())
		if claims == nil {
			api.HandleError(w, api.ErrUnauthorized)
			return
		}

		agentIDStr := chi.URLParam(r, "agentID")
		agentID, err := uuid.Parse(agentIDStr)
		if err != nil {
			api.HandleError(w, api.NewBadRequestError("invalid agent ID"))
			return
		}

		agent, err := h.svc.GetByID(r.Context(), agentID)
		if err != nil {
			slog.Error("fetching agent for ownership check", "error", err)
			api.HandleError(w, api.ErrInternalServer)
			return
		}
		if agent == nil {
			api.HandleError(w, api.NewNotFoundError("agent not found"))
			return
		}

		// CRITICAL: Ownership check
		if agent.OwnerUserID.String() != claims.UserID {
			slog.Warn("ownership violation attempt",
				"agent_id", agentID,
				"agent_owner", agent.OwnerUserID,
				"requester", claims.UserID,
				"path", r.URL.Path,
				"method", r.Method,
			)
			api.HandleError(w, api.ErrOwnershipViolation)
			return
		}

		ctx := SetAgentInContext(r.Context(), agent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
