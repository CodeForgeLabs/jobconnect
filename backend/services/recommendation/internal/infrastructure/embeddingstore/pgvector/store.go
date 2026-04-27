// Package pgvector provides a Postgres-backed EmbeddingStore that uses the
// pgvector extension. Freelancer embeddings live in `freelancer_embeddings`
// keyed by user_id (TEXT) and job embeddings live in `job_embeddings` keyed
// by job_id (BIGINT). Both tables store a 384-dim vector and a text hash for
// dedup.
package pgvector

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jobconnect/recommendation/internal/application"
)

// NewPool builds the pgxpool with the same defaults the other services use.
// The caller owns the pool and must Close it when done.
func NewPool(ctx context.Context, postgresURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute
	return pgxpool.NewWithConfig(ctx, cfg)
}

// Store implements application.EmbeddingStore against pgvector tables.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Get(ctx context.Context, sourceType application.EmbeddingSourceType, sourceID string) (application.StoredEmbedding, bool, error) {
	table, idCol, idVal, err := resolveTarget(sourceType, sourceID)
	if err != nil {
		return application.StoredEmbedding{}, false, err
	}

	var (
		hash    string
		vecText string
	)
	query := fmt.Sprintf("SELECT text_hash, embedding::text FROM %s WHERE %s = $1", table, idCol)
	if err := s.pool.QueryRow(ctx, query, idVal).Scan(&hash, &vecText); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return application.StoredEmbedding{}, false, nil
		}
		return application.StoredEmbedding{}, false, fmt.Errorf("pgvector get: %w", err)
	}

	vec, err := decodeVector(vecText)
	if err != nil {
		return application.StoredEmbedding{}, false, err
	}
	return application.StoredEmbedding{
		SourceType: sourceType,
		SourceID:   sourceID,
		TextHash:   hash,
		Vector:     vec,
	}, true, nil
}

func (s *Store) Upsert(ctx context.Context, e application.StoredEmbedding) error {
	table, idCol, idVal, err := resolveTarget(e.SourceType, e.SourceID)
	if err != nil {
		return err
	}
	if len(e.Vector) == 0 {
		return fmt.Errorf("pgvector upsert: empty vector for %s/%s", e.SourceType, e.SourceID)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s, text_hash, embedding, updated_at)
		VALUES ($1, $2, $3::vector, NOW())
		ON CONFLICT (%s) DO UPDATE SET
			text_hash = EXCLUDED.text_hash,
			embedding = EXCLUDED.embedding,
			updated_at = NOW()
	`, table, idCol, idCol)
	if _, err := s.pool.Exec(ctx, query, idVal, e.TextHash, encodeVector(e.Vector)); err != nil {
		return fmt.Errorf("pgvector upsert: %w", err)
	}
	return nil
}

func resolveTarget(sourceType application.EmbeddingSourceType, sourceID string) (table, idCol string, idVal any, err error) {
	switch sourceType {
	case application.EmbeddingSourceTypeFreelancer:
		if strings.TrimSpace(sourceID) == "" {
			return "", "", nil, fmt.Errorf("pgvector: freelancer source id is empty")
		}
		return "freelancer_embeddings", "user_id", sourceID, nil
	case application.EmbeddingSourceTypeJob:
		id, parseErr := strconv.ParseInt(sourceID, 10, 64)
		if parseErr != nil {
			return "", "", nil, fmt.Errorf("pgvector: job source id %q must be int64: %w", sourceID, parseErr)
		}
		return "job_embeddings", "job_id", id, nil
	default:
		return "", "", nil, fmt.Errorf("pgvector: unsupported source type %q", sourceType)
	}
}

func encodeVector(v []float32) string {
	var b strings.Builder
	b.Grow(len(v)*8 + 2)
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

func decodeVector(s string) ([]float32, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("pgvector: invalid vector text %q", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return []float32{}, nil
	}
	parts := strings.Split(inner, ",")
	out := make([]float32, len(parts))
	for i, p := range parts {
		x, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return nil, fmt.Errorf("pgvector: parse component %d: %w", i, err)
		}
		out[i] = float32(x)
	}
	return out, nil
}
