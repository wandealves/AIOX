package memory

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
)

// Handler handles memory HTTP endpoints.
type Handler struct {
	svc      *Service
	validate *validator.Validate
}

// NewHandler creates a new memory handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:      svc,
		validate: validator.New(),
	}
}

// List returns paginated memories for an agent.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	memories, totalCount, err := h.svc.List(r.Context(), agent.ID, agent.OwnerUserID, page, pageSize)
	if err != nil {
		slog.Error("listing memories", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONPaginated(w, http.StatusOK, memories, totalCount, page, pageSize)
}

// Create creates a new memory for an agent.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	var req CreateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	mem, err := h.svc.Create(r.Context(), agent.ID, agent.OwnerUserID, &req)
	if err != nil {
		slog.Error("creating memory", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusCreated, mem)
}

// Search performs a similarity search on agent memories.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	var req SearchMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	results, err := h.svc.Search(r.Context(), agent.ID, agent.OwnerUserID, &req)
	if err != nil {
		slog.Error("searching memories", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, results)
}

// Delete deletes a single memory.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	memoryIDStr := chi.URLParam(r, "memoryID")
	memoryID, err := uuid.Parse(memoryIDStr)
	if err != nil {
		api.HandleError(w, api.NewBadRequestError("invalid memory ID"))
		return
	}

	if err := h.svc.Delete(r.Context(), memoryID, agent.OwnerUserID); err != nil {
		if err.Error() == "memory not found" {
			api.HandleError(w, api.NewNotFoundError("memory not found"))
			return
		}
		slog.Error("deleting memory", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "memory deleted successfully")
}

// DeleteAll deletes all memories for an agent.
func (h *Handler) DeleteAll(w http.ResponseWriter, r *http.Request) {
	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	if err := h.svc.DeleteByAgent(r.Context(), agent.ID, agent.OwnerUserID); err != nil {
		slog.Error("deleting all memories", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "all memories deleted successfully")
}
