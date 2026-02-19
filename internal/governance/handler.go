package governance

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
	"github.com/aiox-platform/aiox/internal/governance/audit"
	"github.com/aiox-platform/aiox/internal/governance/quota"
)

// Handler provides HTTP handlers for governance endpoints.
type Handler struct {
	quotaSvc  *quota.Service
	auditRepo *audit.Repository
}

// NewHandler creates a new governance Handler.
func NewHandler(quotaSvc *quota.Service, auditRepo *audit.Repository) *Handler {
	return &Handler{
		quotaSvc:  quotaSvc,
		auditRepo: auditRepo,
	}
}

// GetQuota returns the authenticated user's current quota status.
func (h *Handler) GetQuota(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	status, err := h.quotaSvc.GetQuota(r.Context(), userID)
	if err != nil {
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, status)
}

// ListAuditLogs returns paginated audit logs for the authenticated user.
func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	params := parseAuditParams(r)

	logs, total, err := h.auditRepo.ListByOwner(r.Context(), userID, params)
	if err != nil {
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONPaginated(w, http.StatusOK, logs, total, params.Page, params.PageSize)
}

// ListAgentAuditLogs returns paginated audit logs for a specific agent.
// Expects the agent to be set in context by the OwnershipMiddleware.
func (h *Handler) ListAgentAuditLogs(w http.ResponseWriter, r *http.Request) {
	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	params := parseAuditParams(r)

	logs, total, err := h.auditRepo.ListByResource(r.Context(), agent.OwnerUserID, agent.ID, params)
	if err != nil {
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONPaginated(w, http.StatusOK, logs, total, params.Page, params.PageSize)
}

func parseAuditParams(r *http.Request) audit.ListParams {
	params := audit.DefaultListParams()

	if et := r.URL.Query().Get("event_type"); et != "" {
		params.EventType = et
	}
	if sev := r.URL.Query().Get("severity"); sev != "" {
		params.Severity = sev
	}
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
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			params.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			params.To = &t
		}
	}

	return params
}
