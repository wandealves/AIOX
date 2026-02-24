package attachments

import (
	"time"

	"github.com/google/uuid"
)

// Attachment represents an uploaded file or image.
type Attachment struct {
	ID          uuid.UUID `json:"id"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	StoragePath string    `json:"storage_path"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
}

// AllowedContentTypes defines supported upload MIME types.
var AllowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
	"text/plain":      true,
	"text/csv":        true,
	"application/json": true,
}

// IsAllowedContentType checks if a MIME type is supported.
func IsAllowedContentType(ct string) bool {
	return AllowedContentTypes[ct]
}

// MaxUploadSize is the default maximum file size (10 MB).
const MaxUploadSize = 10 * 1024 * 1024
