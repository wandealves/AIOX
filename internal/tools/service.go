package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/auth"
)

// Service provides business logic for tool definitions.
type Service struct {
	repo      *Repository
	encryptor *auth.Encryptor
}

// NewService creates a new tools Service.
// encryptor is used to encrypt auth_config before storage.
func NewService(repo *Repository, encryptionKey string) *Service {
	var enc *auth.Encryptor
	if encryptionKey != "" {
		enc, _ = auth.NewEncryptor(encryptionKey)
	}
	return &Service{repo: repo, encryptor: enc}
}

// Create creates a new tool definition for an agent.
func (s *Service) Create(ctx context.Context, agentID uuid.UUID, req CreateToolRequest) (*ToolDefinition, error) {
	now := time.Now().UTC()

	params := req.Parameters
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	headers := req.Headers
	if len(headers) == 0 {
		headers = json.RawMessage(`{}`)
	}

	authConfig := req.AuthConfig
	if len(authConfig) == 0 {
		authConfig = json.RawMessage(`{}`)
	}

	// Encrypt auth_config if encryptor is available
	if s.encryptor != nil && len(authConfig) > 2 { // > "{}"
		encrypted, err := s.encryptor.Encrypt(string(authConfig))
		if err == nil {
			authConfig = json.RawMessage(`"` + encrypted + `"`)
		}
	}

	timeoutSec := req.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	tool := &ToolDefinition{
		ID:          uuid.New(),
		AgentID:     agentID,
		Name:        req.Name,
		Description: req.Description,
		Parameters:  params,
		EndpointURL: req.EndpointURL,
		HTTPMethod:  req.HTTPMethod,
		Headers:     headers,
		AuthType:    req.AuthType,
		AuthConfig:  authConfig,
		IsActive:    true,
		TimeoutSec:  timeoutSec,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, tool); err != nil {
		return nil, err
	}
	return tool, nil
}

// GetByID returns a tool definition by ID.
func (s *Service) GetByID(ctx context.Context, toolID uuid.UUID) (*ToolDefinition, error) {
	return s.repo.GetByID(ctx, toolID)
}

// ListByAgent returns all tools for an agent.
func (s *Service) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]*ToolDefinition, error) {
	return s.repo.ListByAgent(ctx, agentID)
}

// ListActiveByAgent returns only active tools for an agent.
func (s *Service) ListActiveByAgent(ctx context.Context, agentID uuid.UUID) ([]*ToolDefinition, error) {
	return s.repo.ListActiveByAgent(ctx, agentID)
}

// Update updates an existing tool definition.
func (s *Service) Update(ctx context.Context, toolID uuid.UUID, req UpdateToolRequest) (*ToolDefinition, error) {
	tool, err := s.repo.GetByID(ctx, toolID)
	if err != nil || tool == nil {
		return nil, err
	}

	if req.Name != nil {
		tool.Name = *req.Name
	}
	if req.Description != nil {
		tool.Description = *req.Description
	}
	if req.Parameters != nil {
		tool.Parameters = *req.Parameters
	}
	if req.EndpointURL != nil {
		tool.EndpointURL = *req.EndpointURL
	}
	if req.HTTPMethod != nil {
		tool.HTTPMethod = *req.HTTPMethod
	}
	if req.Headers != nil {
		tool.Headers = *req.Headers
	}
	if req.AuthType != nil {
		tool.AuthType = *req.AuthType
	}
	if req.AuthConfig != nil {
		authConfig := *req.AuthConfig
		if s.encryptor != nil && len(authConfig) > 2 {
			encrypted, encErr := s.encryptor.Encrypt(string(authConfig))
			if encErr == nil {
				authConfig = json.RawMessage(`"` + encrypted + `"`)
			}
		}
		tool.AuthConfig = authConfig
	}
	if req.IsActive != nil {
		tool.IsActive = *req.IsActive
	}
	if req.TimeoutSec != nil {
		tool.TimeoutSec = *req.TimeoutSec
	}
	tool.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, tool); err != nil {
		return nil, err
	}
	return tool, nil
}

// Delete deletes a tool definition.
func (s *Service) Delete(ctx context.Context, toolID uuid.UUID) error {
	return s.repo.Delete(ctx, toolID)
}

// ToolBelongsToAgent checks if a tool belongs to a specific agent.
func (s *Service) ToolBelongsToAgent(ctx context.Context, toolID, agentID uuid.UUID) (bool, error) {
	return s.repo.ToolBelongsToAgent(ctx, toolID, agentID)
}
