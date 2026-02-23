package orgs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// -- Organizations --

func (r *Repository) CreateOrg(ctx context.Context, org *Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, description, settings, quota_config, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Name, org.Slug, org.Description,
		org.Settings, org.QuotaConfig, org.CreatedBy,
		org.CreatedAt, org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating organization: %w", err)
	}
	return nil
}

func (r *Repository) GetOrg(ctx context.Context, id uuid.UUID) (*Organization, error) {
	query := `
		SELECT id, name, slug, description, settings, quota_config, created_by, created_at, updated_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL`
	org := &Organization{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description,
		&org.Settings, &org.QuotaConfig, &org.CreatedBy,
		&org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting organization: %w", err)
	}
	return org, nil
}

func (r *Repository) GetOrgBySlug(ctx context.Context, slug string) (*Organization, error) {
	query := `
		SELECT id, name, slug, description, settings, quota_config, created_by, created_at, updated_at
		FROM organizations WHERE slug = $1 AND deleted_at IS NULL`
	org := &Organization{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description,
		&org.Settings, &org.QuotaConfig, &org.CreatedBy,
		&org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting organization by slug: %w", err)
	}
	return org, nil
}

func (r *Repository) UpdateOrg(ctx context.Context, org *Organization) error {
	query := `
		UPDATE organizations SET name = $2, description = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.pool.Exec(ctx, query, org.ID, org.Name, org.Description, time.Now())
	if err != nil {
		return fmt.Errorf("updating organization: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("organization not found")
	}
	return nil
}

func (r *Repository) SoftDeleteOrg(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE organizations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting organization: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("organization not found")
	}
	return nil
}

func (r *Repository) ListOrgsByUser(ctx context.Context, userID uuid.UUID) ([]*Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.description, o.settings, o.quota_config, o.created_by, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN org_members m ON o.id = m.org_id
		WHERE m.user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.name`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("listing organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org := &Organization{}
		if err := rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.Description,
			&org.Settings, &org.QuotaConfig, &org.CreatedBy,
			&org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning organization: %w", err)
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

// -- Members --

func (r *Repository) AddMember(ctx context.Context, m *Member) error {
	query := `
		INSERT INTO org_members (id, org_id, user_id, role, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, m.ID, m.OrgID, m.UserID, m.Role, m.JoinedAt, m.JoinedAt)
	if err != nil {
		return fmt.Errorf("adding member: %w", err)
	}
	return nil
}

func (r *Repository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*Member, error) {
	query := `
		SELECT m.id, m.org_id, m.user_id, m.role, m.joined_at, COALESCE(u.email, '')
		FROM org_members m
		LEFT JOIN users u ON u.id = m.user_id
		WHERE m.org_id = $1 AND m.user_id = $2`
	m := &Member{}
	err := r.pool.QueryRow(ctx, query, orgID, userID).Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.JoinedAt, &m.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting member: %w", err)
	}
	return m, nil
}

func (r *Repository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*Member, error) {
	query := `
		SELECT m.id, m.org_id, m.user_id, m.role, m.joined_at, COALESCE(u.email, '')
		FROM org_members m
		LEFT JOIN users u ON u.id = m.user_id
		WHERE m.org_id = $1
		ORDER BY m.joined_at`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer rows.Close()

	var members []*Member
	for rows.Next() {
		m := &Member{}
		if err := rows.Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.JoinedAt, &m.Email); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	query := `UPDATE org_members SET role = $3, updated_at = NOW() WHERE org_id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, orgID, userID, role)
	if err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

func (r *Repository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	query := `DELETE FROM org_members WHERE org_id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// -- Invites --

func (r *Repository) CreateInvite(ctx context.Context, inv *Invite) error {
	query := `
		INSERT INTO org_invites (id, org_id, email, role, token, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query,
		inv.ID, inv.OrgID, inv.Email, inv.Role, inv.Token,
		inv.InvitedBy, inv.ExpiresAt, inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating invite: %w", err)
	}
	return nil
}

func (r *Repository) GetInviteByToken(ctx context.Context, token string) (*Invite, error) {
	query := `
		SELECT id, org_id, email, role, token, invited_by, accepted_at, expires_at, created_at
		FROM org_invites WHERE token = $1`
	inv := &Invite{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.Token,
		&inv.InvitedBy, &inv.AcceptedAt, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting invite by token: %w", err)
	}
	return inv, nil
}

func (r *Repository) ListInvites(ctx context.Context, orgID uuid.UUID) ([]*Invite, error) {
	query := `
		SELECT id, org_id, email, role, token, invited_by, accepted_at, expires_at, created_at
		FROM org_invites WHERE org_id = $1
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing invites: %w", err)
	}
	defer rows.Close()

	var invites []*Invite
	for rows.Next() {
		inv := &Invite{}
		if err := rows.Scan(
			&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.Token,
			&inv.InvitedBy, &inv.AcceptedAt, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invite: %w", err)
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func (r *Repository) AcceptInvite(ctx context.Context, token string) error {
	query := `UPDATE org_invites SET accepted_at = NOW() WHERE token = $1 AND accepted_at IS NULL`
	result, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("accepting invite: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("invite not found or already accepted")
	}
	return nil
}

func (r *Repository) DeleteInvite(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM org_invites WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting invite: %w", err)
	}
	return nil
}

// -- Agents (org-scoped) --

func (r *Repository) ListAgentsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	query := `
		SELECT id FROM agents
		WHERE org_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing org agents: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning agent id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) CountAgentsByOrg(ctx context.Context, orgID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM agents WHERE org_id = $1 AND deleted_at IS NULL`
	var count int64
	if err := r.pool.QueryRow(ctx, query, orgID).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting org agents: %w", err)
	}
	return count, nil
}
