package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conversation represents a single execution record returned as a chat message.
type Conversation struct {
	ID           uuid.UUID `json:"id"`
	AgentID      uuid.UUID `json:"agent_id"`
	Input        string    `json:"input"`
	Output       string    `json:"output"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	TokensUsed   int       `json:"tokens_used"`
	DurationMs   int       `json:"duration_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

// Repository handles DB queries for chat conversations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new chat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListByAgentAndOwner returns paginated execution records for a given agent and owner.
func (r *Repository) ListByAgentAndOwner(ctx context.Context, agentID, ownerUserID uuid.UUID, limit, offset int) ([]*Conversation, error) {
	query := `
		SELECT id, agent_id, input, output, status, error_message, tokens_used, duration_ms, created_at
		FROM executions
		WHERE agent_id = $1 AND owner_user_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, agentID, ownerUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing conversations: %w", err)
	}
	defer rows.Close()

	var convos []*Conversation
	for rows.Next() {
		c := &Conversation{}
		if err := rows.Scan(&c.ID, &c.AgentID, &c.Input, &c.Output, &c.Status, &c.ErrorMessage, &c.TokensUsed, &c.DurationMs, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning conversation row: %w", err)
		}
		convos = append(convos, c)
	}
	return convos, rows.Err()
}

// CountByAgentAndOwner returns the total count of executions for a given agent and owner.
func (r *Repository) CountByAgentAndOwner(ctx context.Context, agentID, ownerUserID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM executions WHERE agent_id = $1 AND owner_user_id = $2`

	var count int64
	if err := r.pool.QueryRow(ctx, query, agentID, ownerUserID).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting conversations: %w", err)
	}
	return count, nil
}
