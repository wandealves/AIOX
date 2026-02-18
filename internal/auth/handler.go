package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/users"
)

type Handler struct {
	authSvc  *Service
	userSvc  *users.Service
	validate *validator.Validate
}

func NewHandler(authSvc *Service, userSvc *users.Service) *Handler {
	return &Handler{
		authSvc:  authSvc,
		userSvc:  userSvc,
		validate: validator.New(),
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	// Check if email exists
	exists, err := h.userSvc.ExistsByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Error("checking email existence", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}
	if exists {
		api.HandleError(w, api.ErrEmailAlreadyExists)
		return
	}

	// Hash password
	hash, err := HashPassword(req.Password)
	if err != nil {
		slog.Error("hashing password", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	// Create user
	user, err := h.userSvc.Create(r.Context(), req.Email, hash)
	if err != nil {
		slog.Error("creating user", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	// Generate tokens
	tokens, err := h.authSvc.GenerateTokens(user.ID.String(), user.Email)
	if err != nil {
		slog.Error("generating tokens", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusCreated, tokens)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	// Find user
	user, err := h.userSvc.GetByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Error("getting user by email", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}
	if user == nil {
		api.HandleError(w, api.ErrInvalidCredentials)
		return
	}

	// Verify password
	if err := ComparePassword(user.PasswordHash, req.Password); err != nil {
		api.HandleError(w, api.ErrInvalidCredentials)
		return
	}

	// Generate tokens
	tokens, err := h.authSvc.GenerateTokens(user.ID.String(), user.Email)
	if err != nil {
		slog.Error("generating tokens", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.HandleError(w, api.ErrBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		api.HandleError(w, api.NewValidationError(err.Error()))
		return
	}

	tokens, err := h.authSvc.RefreshTokens(req.RefreshToken)
	if err != nil {
		slog.Error("refreshing tokens", "error", err)
		api.HandleError(w, api.ErrInvalidToken)
		return
	}

	api.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r.Context())
	if claims == nil {
		api.HandleError(w, api.ErrUnauthorized)
		return
	}

	if err := h.authSvc.Logout(claims.UserID); err != nil {
		slog.Error("logging out", "error", err)
		api.HandleError(w, api.ErrInternalServer)
		return
	}

	api.JSONMessage(w, http.StatusOK, "logged out successfully")
}
