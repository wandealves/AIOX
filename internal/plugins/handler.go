package plugins

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

// Handler handles HTTP requests for plugins.
type Handler struct {
	svc *Service
}

// NewHandler creates a new plugin Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register handles POST /plugins/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.JSONErrorMessage(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req RegisterPluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		api.JSONErrorMessage(w, http.StatusBadRequest, "name is required")
		return
	}

	plugin, err := h.svc.Register(r.Context(), req, userID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.JSON(w, http.StatusCreated, plugin)
}

// List handles GET /plugins
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	plugins, err := h.svc.List(r.Context())
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list plugins")
		return
	}

	if plugins == nil {
		plugins = []*Plugin{}
	}
	api.JSON(w, http.StatusOK, plugins)
}

// Delete handles DELETE /plugins/{pluginID}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	pluginIDStr := chi.URLParam(r, "pluginID")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid plugin ID")
		return
	}

	existing, err := h.svc.GetByID(r.Context(), pluginID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check plugin")
		return
	}
	if existing == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "plugin not found")
		return
	}

	if err := h.svc.Delete(r.Context(), pluginID); err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to delete plugin")
		return
	}

	api.JSONMessage(w, http.StatusOK, "plugin deleted")
}

// Sync handles POST /plugins/{pluginID}/sync
func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	pluginIDStr := chi.URLParam(r, "pluginID")
	pluginID, err := uuid.Parse(pluginIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid plugin ID")
		return
	}

	if err := h.svc.Sync(r.Context(), pluginID); err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.JSONMessage(w, http.StatusOK, "plugin synced")
}
