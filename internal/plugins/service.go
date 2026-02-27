package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/tools"
)

// Service provides business logic for plugins.
type Service struct {
	repo           *Repository
	systemToolRepo *tools.SystemRepository
	httpClient     *http.Client
}

// NewService creates a new plugin service.
func NewService(repo *Repository, systemToolRepo *tools.SystemRepository) *Service {
	return &Service{
		repo:           repo,
		systemToolRepo: systemToolRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Register registers a new plugin and creates its tools.
func (s *Service) Register(ctx context.Context, req RegisterPluginRequest, createdBy uuid.UUID) (*Plugin, error) {
	// Check if plugin name already exists
	existing, err := s.repo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("checking existing plugin: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("plugin with name '%s' already exists", req.Name)
	}

	now := time.Now().UTC()
	plugin := &Plugin{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		ManifestURL: req.ManifestURL,
		IsActive:    true,
		AutoSync:    req.AutoSync,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, plugin); err != nil {
		return nil, err
	}

	// Register tools from inline definitions or manifest
	var manifestTools []ManifestTool
	if len(req.Tools) > 0 {
		manifestTools = req.Tools
	} else if req.ManifestURL != "" {
		manifest, err := s.fetchManifest(req.ManifestURL)
		if err != nil {
			return plugin, fmt.Errorf("fetching manifest: %w", err)
		}
		manifestTools = manifest.Tools
	}

	if err := s.createToolsFromManifest(ctx, plugin.ID, manifestTools); err != nil {
		return plugin, fmt.Errorf("creating tools from manifest: %w", err)
	}

	if len(manifestTools) > 0 {
		_ = s.repo.UpdateLastSyncedAt(ctx, plugin.ID)
	}

	return plugin, nil
}

// List returns all registered plugins.
func (s *Service) List(ctx context.Context) ([]*Plugin, error) {
	return s.repo.List(ctx)
}

// GetByID returns a plugin by ID.
func (s *Service) GetByID(ctx context.Context, pluginID uuid.UUID) (*Plugin, error) {
	return s.repo.GetByID(ctx, pluginID)
}

// Delete removes a plugin and its tools (via CASCADE).
func (s *Service) Delete(ctx context.Context, pluginID uuid.UUID) error {
	return s.repo.Delete(ctx, pluginID)
}

// Sync re-fetches the manifest and updates the plugin's tools.
func (s *Service) Sync(ctx context.Context, pluginID uuid.UUID) error {
	plugin, err := s.repo.GetByID(ctx, pluginID)
	if err != nil || plugin == nil {
		return fmt.Errorf("plugin not found")
	}

	if plugin.ManifestURL == "" {
		return fmt.Errorf("plugin has no manifest URL")
	}

	manifest, err := s.fetchManifest(plugin.ManifestURL)
	if err != nil {
		return fmt.Errorf("fetching manifest: %w", err)
	}

	// Remove old tools and recreate
	if err := s.systemToolRepo.DeleteByPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("deleting old tools: %w", err)
	}

	if err := s.createToolsFromManifest(ctx, pluginID, manifest.Tools); err != nil {
		return fmt.Errorf("creating tools: %w", err)
	}

	return s.repo.UpdateLastSyncedAt(ctx, pluginID)
}

func (s *Service) fetchManifest(url string) (*Manifest, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("reading manifest body: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest JSON: %w", err)
	}

	return &manifest, nil
}

func (s *Service) createToolsFromManifest(ctx context.Context, pluginID uuid.UUID, manifestTools []ManifestTool) error {
	now := time.Now().UTC()
	for _, mt := range manifestTools {
		params := mt.Parameters
		if len(params) == 0 {
			params = json.RawMessage(`{}`)
		}

		outputSchema := mt.OutputSchema
		if len(outputSchema) == 0 {
			outputSchema = json.RawMessage(`{}`)
		}

		category := mt.Category
		if category == "" {
			category = "plugin"
		}

		httpMethod := mt.HTTPMethod
		if httpMethod == "" {
			httpMethod = "POST"
		}

		timeoutSec := mt.TimeoutSec
		if timeoutSec <= 0 {
			timeoutSec = 30
		}

		tool := &tools.SystemTool{
			ID:           uuid.New(),
			Name:         mt.Name,
			Description:  mt.Description,
			Category:     category,
			Parameters:   params,
			OutputSchema: outputSchema,
			ToolType:     "http",
			EndpointURL:  mt.EndpointURL,
			HTTPMethod:   httpMethod,
			Headers:      json.RawMessage(`{}`),
			AuthType:     mt.AuthType,
			AuthConfig:   json.RawMessage(`{}`),
			IsActive:     true,
			TimeoutSec:   timeoutSec,
			Version:      "1.0.0",
			PluginID:     &pluginID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := s.systemToolRepo.Create(ctx, tool); err != nil {
			return fmt.Errorf("creating tool '%s': %w", mt.Name, err)
		}
	}
	return nil
}
