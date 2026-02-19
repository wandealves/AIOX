package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

// Repository defines memory persistence operations.
type Repository interface {
	Create(ctx context.Context, mem *Memory) error
	SearchSimilar(ctx context.Context, agentID, ownerUserID uuid.UUID, embedding []float32, limit int, threshold float64) ([]SearchResult, error)
	ListByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID, page, pageSize int) ([]Memory, error)
	CountByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID) (int64, error)
	GetByID(ctx context.Context, id, ownerUserID uuid.UUID) (*Memory, error)
	Delete(ctx context.Context, id, ownerUserID uuid.UUID) error
	DeleteByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID) error
}

// PostgresRepository implements Repository using pgx + pgvector.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new memory repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, mem *Memory) error {
	if mem.ID == uuid.Nil {
		mem.ID = uuid.New()
	}

	metadataBytes := mem.Metadata
	if len(metadataBytes) == 0 {
		metadataBytes = json.RawMessage(`{}`)
	}

	if len(mem.Embedding) > 0 {
		vec := pgvector.NewVector(mem.Embedding)
		_, err := r.pool.Exec(ctx,
			`INSERT INTO agent_memories (id, owner_user_id, agent_id, content, embedding, memory_type, metadata)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			mem.ID, mem.OwnerUserID, mem.AgentID, mem.Content, vec, mem.MemoryType, metadataBytes,
		)
		if err != nil {
			return fmt.Errorf("inserting memory with embedding: %w", err)
		}
	} else {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO agent_memories (id, owner_user_id, agent_id, content, memory_type, metadata)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			mem.ID, mem.OwnerUserID, mem.AgentID, mem.Content, mem.MemoryType, metadataBytes,
		)
		if err != nil {
			return fmt.Errorf("inserting memory: %w", err)
		}
	}
	return nil
}

func (r *PostgresRepository) SearchSimilar(ctx context.Context, agentID, ownerUserID uuid.UUID, embedding []float32, limit int, threshold float64) ([]SearchResult, error) {
	vec := pgvector.NewVector(embedding)
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, agent_id, content, memory_type, metadata, created_at,
		        1 - (embedding <=> $1) AS similarity
		 FROM agent_memories
		 WHERE agent_id = $2 AND owner_user_id = $3
		   AND embedding IS NOT NULL
		   AND 1 - (embedding <=> $1) >= $4
		 ORDER BY embedding <=> $1
		 LIMIT $5`,
		vec, agentID, ownerUserID, threshold, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("searching similar memories: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var m Memory
		var similarity float64
		if err := rows.Scan(&m.ID, &m.OwnerUserID, &m.AgentID, &m.Content, &m.MemoryType, &m.Metadata, &m.CreatedAt, &similarity); err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}
		results = append(results, SearchResult{Memory: m, Similarity: similarity})
	}
	return results, rows.Err()
}

func (r *PostgresRepository) ListByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID, page, pageSize int) ([]Memory, error) {
	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, agent_id, content, memory_type, metadata, created_at
		 FROM agent_memories
		 WHERE agent_id = $1 AND owner_user_id = $2
		 ORDER BY created_at DESC
		 LIMIT $3 OFFSET $4`,
		agentID, ownerUserID, pageSize, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("listing memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.OwnerUserID, &m.AgentID, &m.Content, &m.MemoryType, &m.Metadata, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning memory: %w", err)
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

func (r *PostgresRepository) CountByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM agent_memories WHERE agent_id = $1 AND owner_user_id = $2`,
		agentID, ownerUserID,
	).Scan(&count)
	return count, err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id, ownerUserID uuid.UUID) (*Memory, error) {
	var m Memory
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_user_id, agent_id, content, memory_type, metadata, created_at
		 FROM agent_memories
		 WHERE id = $1 AND owner_user_id = $2`,
		id, ownerUserID,
	).Scan(&m.ID, &m.OwnerUserID, &m.AgentID, &m.Content, &m.MemoryType, &m.Metadata, &m.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("getting memory: %w", err)
	}
	return &m, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id, ownerUserID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM agent_memories WHERE id = $1 AND owner_user_id = $2`,
		id, ownerUserID,
	)
	if err != nil {
		return fmt.Errorf("deleting memory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

func (r *PostgresRepository) DeleteByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM agent_memories WHERE agent_id = $1 AND owner_user_id = $2`,
		agentID, ownerUserID,
	)
	if err != nil {
		return fmt.Errorf("deleting agent memories: %w", err)
	}
	return nil
}
