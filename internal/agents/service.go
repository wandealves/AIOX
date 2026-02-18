package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/auth"
)

type Service struct {
	repo       Repository
	encryptor  *auth.Encryptor
	xmppDomain string
}

func NewService(repo Repository, encryptionKey, xmppDomain string) *Service {
	enc, err := auth.NewEncryptor(encryptionKey)
	if err != nil {
		panic(fmt.Sprintf("failed to create encryptor: %v", err))
	}
	return &Service{
		repo:       repo,
		encryptor:  enc,
		xmppDomain: xmppDomain,
	}
}

func (s *Service) Create(ctx context.Context, ownerID uuid.UUID, req *CreateAgentRequest) (*Agent, error) {
	agentID := uuid.New()
	now := time.Now()

	// Generate JID: agent-<uuid>@agents.<domain>
	jid := fmt.Sprintf("agent-%s@agents.%s", agentID.String(), s.xmppDomain)

	// Encrypt system prompt
	encryptedPrompt, err := s.encryptor.Encrypt(req.SystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("encrypting system prompt: %w", err)
	}

	profile := AgentProfile{
		Name:              req.Name,
		Description:       req.Description,
		SystemPrompt:      encryptedPrompt,
		PersonalityTraits: req.PersonalityTraits,
		Encrypted:         true,
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshaling profile: %w", err)
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = "private"
	}

	row := &AgentRow{
		ID:           agentID,
		OwnerUserID:  ownerID,
		JID:          jid,
		Profile:      profileJSON,
		LLMConfig:    defaultJSON(req.LLMConfig),
		Capabilities: defaultJSON(req.Capabilities),
		MemoryConfig: defaultJSON(req.MemoryConfig),
		Governance:   defaultJSON(req.Governance),
		Visibility:   visibility,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}

	return s.rowToAgent(row)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return s.rowToAgent(row)
}

func (s *Service) ListByOwner(ctx context.Context, ownerID uuid.UUID, params ListAgentsParams) ([]*Agent, int64, error) {
	offset := (params.Page - 1) * params.PageSize

	rows, err := s.repo.ListByOwner(ctx, ownerID, params.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.CountByOwner(ctx, ownerID)
	if err != nil {
		return nil, 0, err
	}

	agents := make([]*Agent, 0, len(rows))
	for _, row := range rows {
		agent, err := s.rowToAgent(row)
		if err != nil {
			return nil, 0, err
		}
		agents = append(agents, agent)
	}

	return agents, count, nil
}

func (s *Service) Update(ctx context.Context, agent *Agent, req *UpdateAgentRequest) (*Agent, error) {
	// Parse current profile
	profile := agent.Profile

	if req.Name != nil {
		profile.Name = *req.Name
	}
	if req.Description != nil {
		profile.Description = *req.Description
	}
	if req.SystemPrompt != nil {
		encrypted, err := s.encryptor.Encrypt(*req.SystemPrompt)
		if err != nil {
			return nil, fmt.Errorf("encrypting system prompt: %w", err)
		}
		profile.SystemPrompt = encrypted
		profile.Encrypted = true
	}
	if req.PersonalityTraits != nil {
		profile.PersonalityTraits = *req.PersonalityTraits
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshaling profile: %w", err)
	}

	visibility := agent.Visibility
	if req.Visibility != nil {
		visibility = *req.Visibility
	}

	llmConfig := agent.LLMConfig
	if req.LLMConfig != nil {
		llmConfig = *req.LLMConfig
	}
	capabilities := agent.Capabilities
	if req.Capabilities != nil {
		capabilities = *req.Capabilities
	}
	memoryConfig := agent.MemoryConfig
	if req.MemoryConfig != nil {
		memoryConfig = *req.MemoryConfig
	}
	governance := agent.Governance
	if req.Governance != nil {
		governance = *req.Governance
	}

	row := &AgentRow{
		ID:           agent.ID,
		OwnerUserID:  agent.OwnerUserID,
		JID:          agent.JID,
		Profile:      profileJSON,
		LLMConfig:    defaultJSON(llmConfig),
		Capabilities: defaultJSON(capabilities),
		MemoryConfig: defaultJSON(memoryConfig),
		Governance:   defaultJSON(governance),
		Visibility:   visibility,
		CreatedAt:    agent.CreatedAt,
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Update(ctx, row); err != nil {
		return nil, err
	}

	return s.rowToAgent(row)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) rowToAgent(row *AgentRow) (*Agent, error) {
	var profile AgentProfile
	if err := json.Unmarshal(row.Profile, &profile); err != nil {
		return nil, fmt.Errorf("unmarshaling profile: %w", err)
	}

	// Decrypt system prompt for the response
	if profile.Encrypted && profile.SystemPrompt != "" {
		decrypted, err := s.encryptor.Decrypt(profile.SystemPrompt)
		if err != nil {
			// If decryption fails, check if it was stored unencrypted
			if !strings.HasPrefix(profile.SystemPrompt, "0") || len(profile.SystemPrompt) < 30 {
				// Likely not encrypted, keep as-is
			} else {
				return nil, fmt.Errorf("decrypting system prompt: %w", err)
			}
		} else {
			profile.SystemPrompt = decrypted
		}
	}

	return &Agent{
		ID:           row.ID,
		OwnerUserID:  row.OwnerUserID,
		JID:          row.JID,
		Profile:      profile,
		LLMConfig:    row.LLMConfig,
		Capabilities: row.Capabilities,
		MemoryConfig: row.MemoryConfig,
		Governance:   row.Governance,
		Visibility:   row.Visibility,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    row.DeletedAt,
	}, nil
}

func defaultJSON(data json.RawMessage) []byte {
	if len(data) == 0 {
		return []byte("{}")
	}
	return data
}
