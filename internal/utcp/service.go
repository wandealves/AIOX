package utcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/auth"
	"github.com/aiox-platform/aiox/internal/tools"
)

// Service provides business logic for UTCP manual management.
type Service struct {
	repo       *Repository
	toolsRepo  *tools.Repository
	encryptor  *auth.Encryptor
	httpClient *http.Client
}

// NewService creates a new UTCP service.
func NewService(repo *Repository, toolsRepo *tools.Repository, encryptionKey string) *Service {
	var enc *auth.Encryptor
	if encryptionKey != "" {
		enc, _ = auth.NewEncryptor(encryptionKey)
	}
	return &Service{
		repo:      repo,
		toolsRepo: toolsRepo,
		encryptor: enc,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Discover fetches a UTCP manual URL and returns a preview without storing anything.
func (s *Service) Discover(ctx context.Context, manualURL string) (*DiscoverResult, error) {
	data, err := s.fetchManualURL(manualURL)
	if err != nil {
		return nil, err
	}

	manual, err := ParseManual(data)
	if err != nil {
		return nil, err
	}

	return &DiscoverResult{
		Name:        manual.Name,
		Description: manual.Description,
		Version:     manual.Version,
		AuthVars:    GetAuthVars(manual.Auth),
		Tools:       manual.Tools,
	}, nil
}

// Connect fetches, parses, and stores a UTCP manual + creates agent tools.
func (s *Service) Connect(ctx context.Context, agentID uuid.UUID, req ConnectManualRequest) (*UTCPManual, error) {
	// Check for duplicate name
	existing, err := s.repo.GetByAgentAndName(ctx, agentID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("checking existing manual: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("a UTCP manual named '%s' is already connected to this agent", req.Name)
	}

	data, err := s.fetchManualURL(req.ManualURL)
	if err != nil {
		return nil, err
	}

	manual, err := ParseManual(data)
	if err != nil {
		return nil, err
	}

	// Encrypt auth values
	var encryptedValues string
	if len(req.AuthValues) > 0 {
		valJSON, _ := json.Marshal(req.AuthValues)
		if s.encryptor != nil {
			encrypted, encErr := s.encryptor.Encrypt(string(valJSON))
			if encErr == nil {
				encryptedValues = encrypted
			} else {
				encryptedValues = string(valJSON)
			}
		} else {
			encryptedValues = string(valJSON)
		}
	}

	authTemplate := json.RawMessage(`{}`)
	if manual.Auth != nil {
		at, _ := json.Marshal(manual.Auth)
		authTemplate = at
	}

	now := time.Now().UTC()
	manualRow := &UTCPManual{
		ID:           uuid.New(),
		AgentID:      agentID,
		Name:         req.Name,
		Description:  manual.Description,
		ManualURL:    req.ManualURL,
		UTCPVersion:  manual.Version,
		AuthTemplate: authTemplate,
		AuthValues:   encryptedValues,
		IsActive:     true,
		LastSyncedAt: &now,
		ToolsCount:   len(manual.Tools),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, manualRow); err != nil {
		return nil, fmt.Errorf("storing UTCP manual: %w", err)
	}

	// Create agent tools
	toolReqs := ConvertToToolRequests(manual, req.Name, req.AuthValues)
	if err := s.createTools(ctx, agentID, manualRow.ID, toolReqs); err != nil {
		return nil, fmt.Errorf("creating tools: %w", err)
	}

	return manualRow, nil
}

// Sync re-fetches a UTCP manual and refreshes its tools.
func (s *Service) Sync(ctx context.Context, manualID uuid.UUID) (*UTCPManual, error) {
	m, err := s.repo.GetByID(ctx, manualID)
	if err != nil || m == nil {
		return nil, fmt.Errorf("manual not found")
	}

	data, err := s.fetchManualURL(m.ManualURL)
	if err != nil {
		return nil, err
	}

	manual, err := ParseManual(data)
	if err != nil {
		return nil, err
	}

	// Decrypt stored auth values
	authValues := s.decryptAuthValues(m.AuthValues)

	// Delete old tools
	if err := s.repo.DeleteToolsByManual(ctx, manualID); err != nil {
		return nil, fmt.Errorf("deleting old tools: %w", err)
	}

	// Recreate tools
	toolReqs := ConvertToToolRequests(manual, m.Name, authValues)
	if err := s.createTools(ctx, m.AgentID, manualID, toolReqs); err != nil {
		return nil, fmt.Errorf("creating tools: %w", err)
	}

	if err := s.repo.UpdateSyncInfo(ctx, manualID, len(manual.Tools)); err != nil {
		return nil, fmt.Errorf("updating sync info: %w", err)
	}

	return s.repo.GetByID(ctx, manualID)
}

// GetByID returns a UTCP manual by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*UTCPManual, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByAgent returns all UTCP manuals for an agent.
func (s *Service) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]*UTCPManual, error) {
	return s.repo.ListByAgent(ctx, agentID)
}

// Delete removes a UTCP manual and its tools (via cascade).
func (s *Service) Delete(ctx context.Context, manualID uuid.UUID) error {
	return s.repo.Delete(ctx, manualID)
}

// ManualBelongsToAgent checks ownership.
func (s *Service) ManualBelongsToAgent(ctx context.Context, manualID, agentID uuid.UUID) (bool, error) {
	return s.repo.ManualBelongsToAgent(ctx, manualID, agentID)
}

func (s *Service) fetchManualURL(url string) ([]byte, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching UTCP manual: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("UTCP manual returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("reading UTCP manual body: %w", err)
	}
	return data, nil
}

func (s *Service) createTools(ctx context.Context, agentID, manualID uuid.UUID, reqs []tools.CreateToolRequest) error {
	now := time.Now().UTC()
	for _, req := range reqs {
		authConfig := req.AuthConfig
		if s.encryptor != nil && len(authConfig) > 2 {
			encrypted, encErr := s.encryptor.Encrypt(string(authConfig))
			if encErr == nil {
				authConfig = json.RawMessage(`"` + encrypted + `"`)
			}
		}

		tool := &tools.ToolDefinition{
			ID:           uuid.New(),
			AgentID:      agentID,
			Name:         req.Name,
			Description:  req.Description,
			Parameters:   req.Parameters,
			EndpointURL:  req.EndpointURL,
			HTTPMethod:   req.HTTPMethod,
			Headers:      req.Headers,
			AuthType:     req.AuthType,
			AuthConfig:   authConfig,
			IsActive:     true,
			TimeoutSec:   req.TimeoutSec,
			Version:      "1.0.0",
			Category:     "utcp",
			OutputSchema: json.RawMessage(`{}`),
			ToolType:     "utcp",
			UTCPManualID: &manualID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := s.toolsRepo.Create(ctx, tool); err != nil {
			return fmt.Errorf("creating tool '%s': %w", req.Name, err)
		}
	}
	return nil
}

func (s *Service) decryptAuthValues(encrypted string) map[string]string {
	if encrypted == "" {
		return nil
	}

	raw := encrypted
	if s.encryptor != nil {
		decrypted, err := s.encryptor.Decrypt(encrypted)
		if err == nil {
			raw = decrypted
		}
	}

	var values map[string]string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return values
}
