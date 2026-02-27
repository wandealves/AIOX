package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/auth"
)

// SystemService provides business logic for system tool definitions.
type SystemService struct {
	repo      *SystemRepository
	encryptor *auth.Encryptor
}

// NewSystemService creates a new system tools Service.
func NewSystemService(repo *SystemRepository, encryptionKey string) *SystemService {
	var enc *auth.Encryptor
	if encryptionKey != "" {
		enc, _ = auth.NewEncryptor(encryptionKey)
	}
	return &SystemService{repo: repo, encryptor: enc}
}

// Create creates a new system tool.
func (s *SystemService) Create(ctx context.Context, req CreateSystemToolRequest) (*SystemTool, error) {
	now := time.Now().UTC()

	params := req.Parameters
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	outputSchema := req.OutputSchema
	if len(outputSchema) == 0 {
		outputSchema = json.RawMessage(`{}`)
	}

	headers := req.Headers
	if len(headers) == 0 {
		headers = json.RawMessage(`{}`)
	}

	authConfig := req.AuthConfig
	if len(authConfig) == 0 {
		authConfig = json.RawMessage(`{}`)
	}

	if s.encryptor != nil && len(authConfig) > 2 {
		encrypted, err := s.encryptor.Encrypt(string(authConfig))
		if err == nil {
			authConfig = json.RawMessage(`"` + encrypted + `"`)
		}
	}

	toolType := req.ToolType
	if toolType == "" {
		toolType = "builtin"
	}

	category := req.Category
	if category == "" {
		category = "utilities"
	}

	httpMethod := req.HTTPMethod
	if httpMethod == "" {
		httpMethod = "GET"
	}

	timeoutSec := req.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	version := req.Version
	if version == "" {
		version = "1.0.0"
	}

	tool := &SystemTool{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		Category:     category,
		Parameters:   params,
		OutputSchema: outputSchema,
		ToolType:     toolType,
		EndpointURL:  req.EndpointURL,
		HTTPMethod:   httpMethod,
		Headers:      headers,
		AuthType:     req.AuthType,
		AuthConfig:   authConfig,
		IsActive:     true,
		TimeoutSec:   timeoutSec,
		Version:      version,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, tool); err != nil {
		return nil, err
	}
	return tool, nil
}

// GetByID returns a system tool by ID.
func (s *SystemService) GetByID(ctx context.Context, toolID uuid.UUID) (*SystemTool, error) {
	return s.repo.GetByID(ctx, toolID)
}

// GetByName returns a system tool by name.
func (s *SystemService) GetByName(ctx context.Context, name string) (*SystemTool, error) {
	return s.repo.GetByName(ctx, name)
}

// List returns all system tools.
func (s *SystemService) List(ctx context.Context) ([]*SystemTool, error) {
	return s.repo.List(ctx)
}

// ListActive returns only active system tools.
func (s *SystemService) ListActive(ctx context.Context) ([]*SystemTool, error) {
	return s.repo.ListActive(ctx)
}

// Update updates an existing system tool.
func (s *SystemService) Update(ctx context.Context, toolID uuid.UUID, req UpdateSystemToolRequest) (*SystemTool, error) {
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
	if req.Category != nil {
		tool.Category = *req.Category
	}
	if req.Parameters != nil {
		tool.Parameters = *req.Parameters
	}
	if req.OutputSchema != nil {
		tool.OutputSchema = *req.OutputSchema
	}
	if req.ToolType != nil {
		tool.ToolType = *req.ToolType
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
	if req.Version != nil {
		tool.Version = *req.Version
	}
	tool.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, tool); err != nil {
		return nil, err
	}
	return tool, nil
}

// Delete deletes a system tool.
func (s *SystemService) Delete(ctx context.Context, toolID uuid.UUID) error {
	return s.repo.Delete(ctx, toolID)
}
