package attachments

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

// Handler handles HTTP requests for attachments.
type Handler struct {
	repo    *Repository
	storage Storage
	maxSize int64
}

// NewHandler creates a new attachment handler.
func NewHandler(repo *Repository, storage Storage, maxSizeMB int) *Handler {
	maxSize := int64(maxSizeMB) * 1024 * 1024
	if maxSize <= 0 {
		maxSize = MaxUploadSize
	}
	return &Handler{
		repo:    repo,
		storage: storage,
		maxSize: maxSize,
	}
}

// Upload handles POST /api/v1/attachments (multipart upload).
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
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

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid file upload: "+err.Error())
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !IsAllowedContentType(contentType) {
		api.JSONErrorMessage(w, http.StatusBadRequest, fmt.Sprintf("content type %q not allowed", contentType))
		return
	}

	// Generate unique storage filename
	attachmentID := uuid.New()
	storageFilename := attachmentID.String() + "-" + header.Filename

	storagePath, checksum, sizeBytes, err := h.storage.Save(r.Context(), storageFilename, file)
	if err != nil {
		slog.Error("attachment: saving file", "error", err)
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	attachment := &Attachment{
		ID:          attachmentID,
		OwnerUserID: ownerID,
		Filename:    header.Filename,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		StoragePath: storagePath,
		Checksum:    checksum,
		CreatedAt:   time.Now(),
	}

	if err := h.repo.Create(r.Context(), attachment); err != nil {
		slog.Error("attachment: creating record", "error", err)
		// Clean up stored file
		_ = h.storage.Delete(r.Context(), storagePath)
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to create attachment record")
		return
	}

	api.JSON(w, http.StatusCreated, attachment)
}

// Get handles GET /api/v1/attachments/{attachmentID}
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

	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid attachment ID")
		return
	}

	attachment, err := h.repo.GetByID(r.Context(), attachmentID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if attachment.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	api.JSON(w, http.StatusOK, attachment)
}

// Download handles GET /api/v1/attachments/{attachmentID}/download
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
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

	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid attachment ID")
		return
	}

	attachment, err := h.repo.GetByID(r.Context(), attachmentID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if attachment.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	reader, err := h.storage.Read(r.Context(), attachment.StoragePath)
	if err != nil {
		slog.Error("attachment: reading file", "error", err)
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to read file")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, attachment.Filename))
	io.Copy(w, reader)
}

// List handles GET /api/v1/attachments
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

	atts, err := h.repo.ListByOwner(r.Context(), ownerID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list attachments")
		return
	}
	if atts == nil {
		atts = []Attachment{}
	}

	api.JSON(w, http.StatusOK, atts)
}

// Delete handles DELETE /api/v1/attachments/{attachmentID}
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

	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid attachment ID")
		return
	}

	attachment, err := h.repo.GetByID(r.Context(), attachmentID)
	if err != nil {
		api.HandleError(w, api.ErrNotFound)
		return
	}
	if attachment.OwnerUserID != ownerID {
		api.HandleError(w, api.ErrForbidden)
		return
	}

	// Delete file from storage
	if err := h.storage.Delete(r.Context(), attachment.StoragePath); err != nil {
		slog.Warn("attachment: deleting file from storage", "error", err)
	}

	// Delete DB record
	if err := h.repo.Delete(r.Context(), attachmentID); err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to delete attachment")
		return
	}

	api.JSONMessage(w, http.StatusOK, "attachment deleted")
}
