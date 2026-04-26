// Package memory holds an in-process EmbeddingStore. It exists so the
// recommendation service can exercise the lazy-embed + text-hash dedup path
// without a backing database; the pgvector adapter (Phase 4b) will satisfy
// the same port for durable deployments.
package memory

import (
	"context"
	"sync"

	"jobconnect/recommendation/internal/application"
)

// Store is a thread-safe in-memory EmbeddingStore keyed by (sourceType, sourceID).
// Not durable; contents are lost on process restart.
type Store struct {
	mu      sync.RWMutex
	entries map[string]application.StoredEmbedding
}

func New() *Store {
	return &Store{entries: make(map[string]application.StoredEmbedding)}
}

func (s *Store) Get(_ context.Context, sourceType application.EmbeddingSourceType, sourceID string) (application.StoredEmbedding, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key(sourceType, sourceID)]
	if !ok {
		return application.StoredEmbedding{}, false, nil
	}
	return cloneEmbedding(e), true, nil
}

func (s *Store) Upsert(_ context.Context, e application.StoredEmbedding) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key(e.SourceType, e.SourceID)] = cloneEmbedding(e)
	return nil
}

// Len is exposed for tests and metrics; it is not part of the EmbeddingStore
// port.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

func key(sourceType application.EmbeddingSourceType, sourceID string) string {
	return string(sourceType) + ":" + sourceID
}

func cloneEmbedding(e application.StoredEmbedding) application.StoredEmbedding {
	v := make([]float32, len(e.Vector))
	copy(v, e.Vector)
	return application.StoredEmbedding{
		SourceType: e.SourceType,
		SourceID:   e.SourceID,
		TextHash:   e.TextHash,
		Vector:     v,
	}
}
