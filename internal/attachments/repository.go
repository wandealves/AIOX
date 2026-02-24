package attachments

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles attachment persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new attachment repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new attachment record.
func (r *Repository) Create(ctx context.Context, a *Attachment) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO attachments (id, owner_user_id, filename, content_type, size_bytes, storage_path, checksum, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		a.ID, a.OwnerUserID, a.Filename, a.ContentType, a.SizeBytes, a.StoragePath, a.Checksum, a.CreatedAt,
	).Scan(&a.ID)
}

// GetByID returns an attachment by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Attachment, error) {
	var a Attachment
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_user_id, filename, content_type, size_bytes, storage_path, checksum, created_at
		 FROM attachments WHERE id = $1`, id,
	).Scan(&a.ID, &a.OwnerUserID, &a.Filename, &a.ContentType, &a.SizeBytes, &a.StoragePath, &a.Checksum, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListByOwner returns attachments for a user.
func (r *Repository) ListByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]Attachment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, filename, content_type, size_bytes, storage_path, checksum, created_at
		 FROM attachments WHERE owner_user_id = $1
		 ORDER BY created_at DESC LIMIT 100`, ownerUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.OwnerUserID, &a.Filename, &a.ContentType, &a.SizeBytes, &a.StoragePath, &a.Checksum, &a.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// Delete removes an attachment record.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM attachments WHERE id = $1`, id)
	return err
}
