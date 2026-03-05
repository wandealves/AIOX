package utcp

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
)

// Handler handles HTTP requests for UTCP manuals.
type Handler struct {
	svc *Service
}

// NewHandler creates a new UTCP Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// DiscoverTools handles POST /agents/{agentID}/utcp/discover
func (h *Handler) DiscoverTools(w http.ResponseWriter, r *http.Request) {
	var req DiscoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ManualURL == "" {
		api.JSONErrorMessage(w, http.StatusBadRequest, "manual_url is required")
		return
	}

	result, err := h.svc.Discover(r.Context(), req.ManualURL)
	if err != nil {
		slog.Error("UTCP discover failed", "error", err, "url", req.ManualURL)
		api.JSONErrorMessage(w, http.StatusBadGateway, "failed to discover tools: "+err.Error())
		return
	}

	api.JSON(w, http.StatusOK, result)
}

// ConnectManual handles POST /agents/{agentID}/utcp/manuals
func (h *Handler) ConnectManual(w http.ResponseWriter, r *http.Request) {
	agentID, err := parseAgentID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	var req ConnectManualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.ManualURL == "" {
		api.JSONErrorMessage(w, http.StatusBadRequest, "name and manual_url are required")
		return
	}

	manual, err := h.svc.Connect(r.Context(), agentID, req)
	if err != nil {
		slog.Error("UTCP connect failed", "error", err, "agent_id", agentID, "name", req.Name)
		if strings.Contains(err.Error(), "already connected") {
			api.JSONErrorMessage(w, http.StatusConflict, err.Error())
			return
		}
		api.JSONErrorMessage(w, http.StatusBadGateway, "failed to connect UTCP manual: "+err.Error())
		return
	}

	api.JSON(w, http.StatusCreated, manual)
}

// ListManuals handles GET /agents/{agentID}/utcp/manuals
func (h *Handler) ListManuals(w http.ResponseWriter, r *http.Request) {
	agentID, err := parseAgentID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	manuals, err := h.svc.ListByAgent(r.Context(), agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list manuals")
		return
	}

	if manuals == nil {
		manuals = []*UTCPManual{}
	}
	api.JSON(w, http.StatusOK, manuals)
}

// GetManual handles GET /agents/{agentID}/utcp/manuals/{manualID}
func (h *Handler) GetManual(w http.ResponseWriter, r *http.Request) {
	agentID, err := parseAgentID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	manualID, err := parseManualID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid manual ID")
		return
	}

	belongs, err := h.svc.ManualBelongsToAgent(r.Context(), manualID, agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check manual")
		return
	}
	if !belongs {
		api.JSONErrorMessage(w, http.StatusNotFound, "manual not found")
		return
	}

	manual, err := h.svc.GetByID(r.Context(), manualID)
	if err != nil || manual == nil {
		api.JSONErrorMessage(w, http.StatusNotFound, "manual not found")
		return
	}

	api.JSON(w, http.StatusOK, manual)
}

// SyncManual handles POST /agents/{agentID}/utcp/manuals/{manualID}/sync
func (h *Handler) SyncManual(w http.ResponseWriter, r *http.Request) {
	agentID, err := parseAgentID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	manualID, err := parseManualID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid manual ID")
		return
	}

	belongs, err := h.svc.ManualBelongsToAgent(r.Context(), manualID, agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check manual")
		return
	}
	if !belongs {
		api.JSONErrorMessage(w, http.StatusNotFound, "manual not found")
		return
	}

	manual, err := h.svc.Sync(r.Context(), manualID)
	if err != nil {
		slog.Error("UTCP sync failed", "error", err, "manual_id", manualID)
		api.JSONErrorMessage(w, http.StatusBadGateway, "failed to sync manual: "+err.Error())
		return
	}

	api.JSON(w, http.StatusOK, manual)
}

// DeleteManual handles DELETE /agents/{agentID}/utcp/manuals/{manualID}
func (h *Handler) DeleteManual(w http.ResponseWriter, r *http.Request) {
	agentID, err := parseAgentID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	manualID, err := parseManualID(r)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid manual ID")
		return
	}

	belongs, err := h.svc.ManualBelongsToAgent(r.Context(), manualID, agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to check manual")
		return
	}
	if !belongs {
		api.JSONErrorMessage(w, http.StatusNotFound, "manual not found")
		return
	}

	if err := h.svc.Delete(r.Context(), manualID); err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to delete manual")
		return
	}

	api.JSONMessage(w, http.StatusOK, "UTCP manual disconnected")
}

func parseAgentID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "agentID"))
}

func parseManualID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "manualID"))
}
