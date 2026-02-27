package tools

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
)

// SystemHandler handles HTTP requests for system tool definitions (admin-only).
type SystemHandler struct {
	svc *SystemService
}

// NewSystemHandler creates a new system tools Handler.
func NewSystemHandler(svc *SystemService) *SystemHandler {
	return &SystemHandler{svc: svc}
}

// Create handles POST /admin/tools
func (h *SystemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSystemToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		api.JSONErrorMessage(w, http.StatusBadRequest, "name is required")
		return
	}

	tool, err := h.svc.Create(r.Context(), req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to create system tool")
		return
	}

	api.JSON(w, http.StatusCreated, tool)
}

// List handles GET /admin/tools
func (h *SystemHandler) List(w http.ResponseWriter, r *http.Request) {
	tools, err := h.svc.List(r.Context())
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list system tools")
		return
	}

	if tools == nil {
		tools = []*SystemTool{}
	}
	api.JSON(w, http.StatusOK, tools)
}

// Get handles GET /admin/tools/{toolID}
func (h *SystemHandler) Get(w http.ResponseWriter, r *http.Request) {
	toolIDStr := chi.URLParam(r, "toolID")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid tool ID")
		return
	}

	tool, err := h.svc.GetByID(r.Context(), toolID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to get system tool")
		return
	}
	if tool == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "system tool not found")
		return
	}

	api.JSON(w, http.StatusOK, tool)
}

// Update handles PUT /admin/tools/{toolID}
func (h *SystemHandler) Update(w http.ResponseWriter, r *http.Request) {
	toolIDStr := chi.URLParam(r, "toolID")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid tool ID")
		return
	}

	var req UpdateSystemToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tool, err := h.svc.Update(r.Context(), toolID, req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to update system tool")
		return
	}
	if tool == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "system tool not found")
		return
	}

	api.JSON(w, http.StatusOK, tool)
}

// Delete handles DELETE /admin/tools/{toolID}
func (h *SystemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	toolIDStr := chi.URLParam(r, "toolID")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid tool ID")
		return
	}

	existing, err := h.svc.GetByID(r.Context(), toolID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check system tool")
		return
	}
	if existing == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "system tool not found")
		return
	}

	if err := h.svc.Delete(r.Context(), toolID); err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to delete system tool")
		return
	}

	api.JSONMessage(w, http.StatusOK, "system tool deleted")
}
