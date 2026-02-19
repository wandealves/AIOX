package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles audit_logs PostgreSQL operations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new audit Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Insert persists a single audit log entry.
func (r *Repository) Insert(ctx context.Context, log *AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	detailsJSON := log.Details
	if len(detailsJSON) == 0 {
		detailsJSON = json.RawMessage(`{}`)
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_logs (id, owner_user_id, event_type, severity, resource_type, resource_id, details, ip_address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.OwnerUserID, log.EventType, log.Severity, log.ResourceType, log.ResourceID, detailsJSON, log.IPAddress)
	if err != nil {
		return fmt.Errorf("inserting audit log: %w", err)
	}
	return nil
}

// ListByOwner returns paginated audit logs for an owner with optional filters.
func (r *Repository) ListByOwner(ctx context.Context, ownerUserID uuid.UUID, params ListParams) ([]AuditLog, int64, error) {
	return r.list(ctx, ownerUserID, nil, params)
}

// ListByResource returns paginated audit logs for a specific resource owned by the user.
func (r *Repository) ListByResource(ctx context.Context, ownerUserID uuid.UUID, resourceID uuid.UUID, params ListParams) ([]AuditLog, int64, error) {
	return r.list(ctx, ownerUserID, &resourceID, params)
}

func (r *Repository) list(ctx context.Context, ownerUserID uuid.UUID, resourceID *uuid.UUID, params ListParams) ([]AuditLog, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("owner_user_id = $%d", argIdx))
	args = append(args, ownerUserID)
	argIdx++

	if resourceID != nil {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argIdx))
		args = append(args, *resourceID)
		argIdx++
	}

	if params.EventType != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIdx))
		args = append(args, params.EventType)
		argIdx++
	}

	if params.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, params.Severity)
		argIdx++
	}

	if params.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *params.From)
		argIdx++
	}

	if params.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *params.To)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", where)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("counting audit logs: %w", err)
	}

	// Data query
	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(
		`SELECT id, owner_user_id, event_type, severity, resource_type, resource_id, details, ip_address, created_at
		 FROM audit_logs WHERE %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.OwnerUserID, &l.EventType, &l.Severity,
			&l.ResourceType, &l.ResourceID, &l.Details, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning audit log: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, totalCount, nil
}
