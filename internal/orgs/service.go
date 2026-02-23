package orgs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateOrg creates an organization and adds the creator as owner.
func (s *Service) CreateOrg(ctx context.Context, userID uuid.UUID, req *CreateOrgRequest) (*Organization, error) {
	// Check slug uniqueness
	existing, err := s.repo.GetOrgBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("slug already taken")
	}

	now := time.Now()
	org := &Organization{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Settings:    []byte("{}"),
		QuotaConfig: []byte("{}"),
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateOrg(ctx, org); err != nil {
		return nil, err
	}

	// Add creator as owner
	member := &Member{
		ID:       uuid.New(),
		OrgID:    org.ID,
		UserID:   userID,
		Role:     RoleOwner,
		JoinedAt: now,
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("adding owner member: %w", err)
	}

	return org, nil
}

// GetOrg returns an organization by ID.
func (s *Service) GetOrg(ctx context.Context, id uuid.UUID) (*Organization, error) {
	return s.repo.GetOrg(ctx, id)
}

// ListByUser returns all organizations the user is a member of.
func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Organization, error) {
	return s.repo.ListOrgsByUser(ctx, userID)
}

// UpdateOrg updates organization details.
func (s *Service) UpdateOrg(ctx context.Context, org *Organization, req *UpdateOrgRequest) (*Organization, error) {
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = *req.Description
	}
	if err := s.repo.UpdateOrg(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

// DeleteOrg soft-deletes an organization.
func (s *Service) DeleteOrg(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteOrg(ctx, id)
}

// InviteMember creates an invite with a random token, valid for 7 days.
func (s *Service) InviteMember(ctx context.Context, orgID, inviterID uuid.UUID, req *InviteMemberRequest) (*Invite, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	now := time.Now()
	inv := &Invite{
		ID:        uuid.New(),
		OrgID:     orgID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     hex.EncodeToString(tokenBytes),
		InvitedBy: inviterID,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	if err := s.repo.CreateInvite(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

// AcceptInvite processes an invite acceptance: validates token, creates member.
func (s *Service) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) error {
	inv, err := s.repo.GetInviteByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil {
		return fmt.Errorf("invite not found")
	}
	if inv.AcceptedAt != nil {
		return fmt.Errorf("invite already accepted")
	}
	if time.Now().After(inv.ExpiresAt) {
		return fmt.Errorf("invite expired")
	}

	// Check if already a member
	existing, err := s.repo.GetMember(ctx, inv.OrgID, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("already a member of this organization")
	}

	// Add member
	member := &Member{
		ID:       uuid.New(),
		OrgID:    inv.OrgID,
		UserID:   userID,
		Role:     inv.Role,
		JoinedAt: time.Now(),
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		return err
	}

	// Mark invite as accepted
	return s.repo.AcceptInvite(ctx, token)
}

// GetMember returns a member by org and user IDs.
func (s *Service) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*Member, error) {
	return s.repo.GetMember(ctx, orgID, userID)
}

// ListMembers returns all members of an organization.
func (s *Service) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*Member, error) {
	return s.repo.ListMembers(ctx, orgID)
}

// UpdateMemberRole changes a member's role.
func (s *Service) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	return s.repo.UpdateMemberRole(ctx, orgID, userID, role)
}

// RemoveMember removes a member from the organization.
func (s *Service) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, orgID, userID)
}

// ListInvites returns all invites for an organization.
func (s *Service) ListInvites(ctx context.Context, orgID uuid.UUID) ([]*Invite, error) {
	return s.repo.ListInvites(ctx, orgID)
}

// DeleteInvite cancels a pending invite.
func (s *Service) DeleteInvite(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteInvite(ctx, id)
}
