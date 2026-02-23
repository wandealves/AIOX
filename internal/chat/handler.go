package chat

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

// SendMessageRequest is the body for POST /agents/{agentID}/messages.
type SendMessageRequest struct {
	Body string `json:"body"`
}

// SendMessageResponse is returned with 202 Accepted.
type SendMessageResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// Handler handles chat-related HTTP endpoints.
type Handler struct {
	repo      *Repository
	publisher *inats.Publisher
}

// NewHandler creates a new chat Handler.
func NewHandler(repo *Repository, publisher *inats.Publisher) *Handler {
	return &Handler{repo: repo, publisher: publisher}
}

// SendMessage handles POST /api/v1/agents/{agentID}/messages.
// It publishes an InboundMessage with Channel:"ws" for WebSocket delivery.
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	agent := agents.GetAgentFromContext(r.Context())
	if agent == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if req.Body == "" {
		api.HandleError(w, api.NewBadRequestError("body is required"))
		return
	}

	requestID := uuid.New().String()
	fromJID := "ws-" + claims.UserID + "@ws.local"

	inbound := inats.InboundMessage{
		ID:         requestID,
		FromJID:    fromJID,
		ToJID:      agent.JID,
		Body:       req.Body,
		StanzaType: "chat",
		ReceivedAt: time.Now().UTC(),
		Channel:    inats.ChannelWS,
	}

	if err := h.publisher.PublishInboundMessage(r.Context(), inbound); err != nil {
		slog.Error("chat: publishing inbound message", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusAccepted, SendMessageResponse{
		RequestID: requestID,
		Status:    "queued",
	})
}

// ListConversations handles GET /api/v1/agents/{agentID}/conversations.
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
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

	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	page, pageSize := 1, 20
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

	offset := (page - 1) * pageSize

	convos, err := h.repo.ListByAgentAndOwner(r.Context(), agentID, ownerID, pageSize, offset)
	if err != nil {
		slog.Error("chat: listing conversations", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	totalCount, err := h.repo.CountByAgentAndOwner(r.Context(), agentID, ownerID)
	if err != nil {
		slog.Error("chat: counting conversations", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONPaginated(w, http.StatusOK, convos, totalCount, page, pageSize)
}
