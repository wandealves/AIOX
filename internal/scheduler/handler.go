package scheduler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

// Handler handles HTTP requests for scheduled tasks.
type Handler struct {
	svc *Service
}

// NewHandler creates a new scheduler handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /api/v1/schedules
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

	var req CreateScheduledTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	st, err := h.svc.Create(r.Context(), ownerID, req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSON(w, http.StatusCreated, st)
}

// List handles GET /api/v1/schedules
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

	tasks, err := h.svc.ListByOwner(r.Context(), ownerID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}
	if tasks == nil {
		tasks = []ScheduledTask{}
	}

	api.JSON(w, http.StatusOK, tasks)
}

// Get handles GET /api/v1/schedules/{scheduleID}
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

	scheduleID, err := uuid.Parse(chi.URLParam(r, "scheduleID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid schedule ID")
		return
	}

	st, err := h.svc.GetByID(r.Context(), scheduleID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if st.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	api.JSON(w, http.StatusOK, st)
}

// Update handles PUT /api/v1/schedules/{scheduleID}
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

	scheduleID, err := uuid.Parse(chi.URLParam(r, "scheduleID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid schedule ID")
		return
	}

	var req UpdateScheduledTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	st, err := h.svc.Update(r.Context(), scheduleID, ownerID, req)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSON(w, http.StatusOK, st)
}

// Delete handles DELETE /api/v1/schedules/{scheduleID}
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

	scheduleID, err := uuid.Parse(chi.URLParam(r, "scheduleID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid schedule ID")
		return
	}

	if err := h.svc.Delete(r.Context(), scheduleID, ownerID); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	api.JSONMessage(w, http.StatusOK, "schedule deleted")
}
