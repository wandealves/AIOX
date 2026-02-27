package tools

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
)

// Handler handles HTTP requests for tool definitions.
type Handler struct {
	svc *Service
}

// NewHandler creates a new tools Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /agents/{agentID}/tools
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	var req CreateToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.EndpointURL == "" || req.HTTPMethod == "" {
		api.JSONErrorMessage(w, http.StatusBadRequest, "name, endpoint_url, and http_method are required")
		return
	}

	tool, err := h.svc.Create(r.Context(), agentID, req)
	if err != nil {
		slog.Error("failed to create tool", "error", err, "agent_id", agentID, "tool_name", req.Name)
		if strings.Contains(err.Error(), "idx_agent_tools_unique") {
			api.JSONErrorMessage(w, http.StatusConflict, "a tool with this name already exists for this agent")
			return
		}
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to create tool")
		return
	}

	api.JSON(w, http.StatusCreated, tool)
}

// List handles GET /agents/{agentID}/tools
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	tools, err := h.svc.ListByAgent(r.Context(), agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list tools")
		return
	}

	if tools == nil {
		tools = []*ToolDefinition{}
	}
	api.JSON(w, http.StatusOK, tools)
}

// Get handles GET /agents/{agentID}/tools/{toolID}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	toolIDStr := chi.URLParam(r, "toolID")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid tool ID")
		return
	}

	belongs, err := h.svc.ToolBelongsToAgent(r.Context(), toolID, agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check tool")
		return
	}
	if !belongs {
		api.JSONErrorMessage(w, http.StatusNotFound, "tool not found")
		return
	}

	tool, err := h.svc.GetByID(r.Context(), toolID)
	if err != nil || tool == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "tool not found")
		return
	}

	api.JSON(w, http.StatusOK, tool)
}

// Update handles PUT /agents/{agentID}/tools/{toolID}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	toolIDStr := chi.URLParam(r, "toolID")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid tool ID")
		return
	}

	belongs, err := h.svc.ToolBelongsToAgent(r.Context(), toolID, agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check tool")
		return
	}
	if !belongs {
		api.JSONErrorMessage(w, http.StatusNotFound, "tool not found")
		return
	}

	var req UpdateToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tool, err := h.svc.Update(r.Context(), toolID, req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to update tool")
		return
	}
	if tool == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "tool not found")
		return
	}

	api.JSON(w, http.StatusOK, tool)
}

// Delete handles DELETE /agents/{agentID}/tools/{toolID}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	toolIDStr := chi.URLParam(r, "toolID")
	toolID, err := uuid.Parse(toolIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid tool ID")
		return
	}

	belongs, err := h.svc.ToolBelongsToAgent(r.Context(), toolID, agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check tool")
		return
	}
	if !belongs {
		api.JSONErrorMessage(w, http.StatusNotFound, "tool not found")
		return
	}

	if err := h.svc.Delete(r.Context(), toolID); err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to delete tool")
		return
	}

	api.JSONMessage(w, http.StatusOK, "tool deleted")
}
