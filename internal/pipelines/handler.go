package pipelines

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

// Handler handles HTTP requests for pipelines.
type Handler struct {
	svc *Service
}

// NewHandler creates a new pipeline handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /api/v1/pipelines
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreatePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Create(r.Context(), ownerID, req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSON(w, http.StatusCreated, p)
}

// List handles GET /api/v1/pipelines
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

	pipelines, err := h.svc.ListByOwner(r.Context(), ownerID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list pipelines")
		return
	}
	if pipelines == nil {
		pipelines = []Pipeline{}
	}

	api.JSON(w, http.StatusOK, pipelines)
}

// Get handles GET /api/v1/pipelines/{pipelineID}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "pipelineID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	p, err := h.svc.GetByID(r.Context(), pipelineID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if p.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	api.JSON(w, http.StatusOK, p)
}

// Update handles PUT /api/v1/pipelines/{pipelineID}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "pipelineID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	var req UpdatePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Update(r.Context(), pipelineID, ownerID, req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSON(w, http.StatusOK, p)
}

// Delete handles DELETE /api/v1/pipelines/{pipelineID}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "pipelineID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	if err := h.svc.Delete(r.Context(), pipelineID, ownerID); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSONMessage(w, http.StatusOK, "pipeline deleted")
}

// Execute handles POST /api/v1/pipelines/{pipelineID}/execute
func (h *Handler) Execute(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "pipelineID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	var req ExecutePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	exec, err := h.svc.Execute(r.Context(), ownerID, pipelineID, req.Input)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSON(w, http.StatusOK, exec)
}

// ListExecutions handles GET /api/v1/pipelines/{pipelineID}/executions
func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "pipelineID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	// Verify ownership
	p, err := h.svc.GetByID(r.Context(), pipelineID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if p.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	execs, err := h.svc.ListExecutions(r.Context(), pipelineID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list executions")
		return
	}
	if execs == nil {
		execs = []PipelineExecution{}
	}

	api.JSON(w, http.StatusOK, execs)
}

// GetExecution handles GET /api/v1/pipelines/{pipelineID}/executions/{executionID}
func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
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

	executionID, err := uuid.Parse(chi.URLParam(r, "executionID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	exec, err := h.svc.GetExecution(r.Context(), executionID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if exec.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	api.JSON(w, http.StatusOK, exec)
}
