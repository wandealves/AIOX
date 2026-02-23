package orgs

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

// CreateOrg handles POST /api/v1/orgs
func (h *Handler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	org, err := h.svc.CreateOrg(r.Context(), userID, &req)
	if err != nil {
		if err.Error() == "slug already taken" {
			api.HandleError(w, api.NewConflictError("slug already taken"))
			return
		}
		slog.Error("creating organization", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusCreated, org)
}

// ListOrgs handles GET /api/v1/orgs
func (h *Handler) ListOrgs(w http.ResponseWriter, r *http.Request) {
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

	orgs, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		slog.Error("listing organizations", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, orgs)
}

// GetOrg handles GET /api/v1/orgs/{orgID}
func (h *Handler) GetOrg(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	api.JSON(w, http.StatusOK, org)
}

// UpdateOrg handles PUT /api/v1/orgs/{orgID}
func (h *Handler) UpdateOrg(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	var req UpdateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	updated, err := h.svc.UpdateOrg(r.Context(), org, &req)
	if err != nil {
		slog.Error("updating organization", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, updated)
}

// DeleteOrg handles DELETE /api/v1/orgs/{orgID}
func (h *Handler) DeleteOrg(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	if err := h.svc.DeleteOrg(r.Context(), org.ID); err != nil {
		slog.Error("deleting organization", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "organization deleted successfully")
}

// ListMembers handles GET /api/v1/orgs/{orgID}/members
func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	members, err := h.svc.ListMembers(r.Context(), org.ID)
	if err != nil {
		slog.Error("listing members", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, members)
}

// UpdateMemberRole handles PUT /api/v1/orgs/{orgID}/members/{userID}
func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		api.HandleError(w, api.NewBadRequestError("invalid user ID"))
		return
	}

	var body struct {
		Role string `json:"role" validate:"required,oneof=admin editor viewer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}
	if err := h.validate.Struct(body); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	// Cannot change owner role
	target, err := h.svc.GetMember(r.Context(), org.ID, targetUserID)
	if err != nil {
		api.HandleError(w, api.ErrInternalServer)
		return
	}
	if target == nil {
		api.HandleError(w, api.NewNotFoundError("member not found"))
		return
	}
	if target.Role == RoleOwner {
		api.HandleError(w, api.NewBadRequestError("cannot change owner role"))
		return
	}

	if err := h.svc.UpdateMemberRole(r.Context(), org.ID, targetUserID, body.Role); err != nil {
		slog.Error("updating member role", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "role updated")
}

// RemoveMember handles DELETE /api/v1/orgs/{orgID}/members/{userID}
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		api.HandleError(w, api.NewBadRequestError("invalid user ID"))
		return
	}

	// Cannot remove owner
	target, err := h.svc.GetMember(r.Context(), org.ID, targetUserID)
	if err != nil {
		api.HandleError(w, api.ErrInternalServer)
		return
	}
	if target == nil {
		api.HandleError(w, api.NewNotFoundError("member not found"))
		return
	}
	if target.Role == RoleOwner {
		api.HandleError(w, api.NewBadRequestError("cannot remove organization owner"))
		return
	}

	if err := h.svc.RemoveMember(r.Context(), org.ID, targetUserID); err != nil {
		slog.Error("removing member", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "member removed")
}

// CreateInvite handles POST /api/v1/orgs/{orgID}/invites
func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	inviterID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	invite, err := h.svc.InviteMember(r.Context(), org.ID, inviterID, &req)
	if err != nil {
		slog.Error("creating invite", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusCreated, invite)
}

// ListInvites handles GET /api/v1/orgs/{orgID}/invites
func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	invites, err := h.svc.ListInvites(r.Context(), org.ID)
	if err != nil {
		slog.Error("listing invites", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, invites)
}

// DeleteInvite handles DELETE /api/v1/orgs/{orgID}/invites/{inviteID}
func (h *Handler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	inviteID, err := uuid.Parse(chi.URLParam(r, "inviteID"))
	if err != nil {
		api.HandleError(w, api.NewBadRequestError("invalid invite ID"))
		return
	}

	if err := h.svc.DeleteInvite(r.Context(), inviteID); err != nil {
		slog.Error("deleting invite", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "invite cancelled")
}

// AcceptInvite handles POST /api/v1/invites/{token}/accept
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		api.HandleError(w, api.NewBadRequestError("missing invite token"))
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	if err := h.svc.AcceptInvite(r.Context(), token, userID); err != nil {
		switch err.Error() {
		case "invite not found":
			api.HandleError(w, api.NewNotFoundError("invite not found"))
		case "invite already accepted":
			api.HandleError(w, api.NewConflictError("invite already accepted"))
		case "invite expired":
			api.HandleError(w, api.NewBadRequestError("invite has expired"))
		case "already a member of this organization":
			api.HandleError(w, api.NewConflictError("already a member"))
		default:
			slog.Error("accepting invite", "error", err)
			api.HandleError(w, api.ErrInternalServer)
		}
		return
	}

	api.JSONMessage(w, http.StatusOK, "invite accepted")
}

// ListOrgAgents handles GET /api/v1/orgs/{orgID}/agents
func (h *Handler) ListOrgAgents(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
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

	offset := (page - 1) * pageSize

	ids, err := h.svc.repo.ListAgentsByOrg(r.Context(), org.ID, pageSize, offset)
	if err != nil {
		slog.Error("listing org agents", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	count, err := h.svc.repo.CountAgentsByOrg(r.Context(), org.ID)
	if err != nil {
		slog.Error("counting org agents", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONPaginated(w, http.StatusOK, ids, count, page, pageSize)
}

// CreateOrgAgent handles POST /api/v1/orgs/{orgID}/agents
// This is a lightweight endpoint that returns org info for the client to use
// when creating an agent with org_id via the standard agent creation endpoint.
func (h *Handler) CreateOrgAgent(w http.ResponseWriter, r *http.Request) {
	org := GetOrgFromContext(r.Context())
	if org == nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}

	api.JSON(w, http.StatusOK, map[string]any{
		"org_id":   org.ID,
		"org_name": org.Name,
		"message":  "use POST /api/v1/agents with org_id field to create an agent in this organization",
	})
}
