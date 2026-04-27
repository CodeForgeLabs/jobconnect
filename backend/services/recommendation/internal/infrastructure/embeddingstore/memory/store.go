// Package memory holds an in-process EmbeddingStore. It exists so the
// recommendation service can exercise the lazy-embed + text-hash dedup path
// without a backing database; the pgvector adapter (Phase 4b) will satisfy
// the same port for durable deployments.
package memory

import (
	"context"
	"math"
	"sort"
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

// SearchByVector does a linear scan over entries matching the source type,
// computing cosine distance against the query vector, and returns up to k
// nearest neighbours sorted ascending by distance. Vectors with mismatched
// dimensionality are skipped silently — the assumption is that all entries
// in a deployment share one model and therefore one dimensionality. Empty
// store, k <= 0, or empty query vector all yield (nil, nil).
func (s *Store) SearchByVector(_ context.Context, sourceType application.EmbeddingSourceType, vector []float32, k int) ([]application.VectorHit, error) {
	if k <= 0 || len(vector) == 0 {
		return nil, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	hits := make([]application.VectorHit, 0)
	for _, e := range s.entries {
		if e.SourceType != sourceType {
			continue
		}
		if len(e.Vector) != len(vector) {
			continue
		}
		hits = append(hits, application.VectorHit{
			SourceID: e.SourceID,
			Distance: cosineDistance(vector, e.Vector),
		})
	}
	if len(hits) == 0 {
		return nil, nil
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Distance < hits[j].Distance })
	if k < len(hits) {
		hits = hits[:k]
	}
	return hits, nil
}

func cosineDistance(a, b []float32) float32 {
	var dot, na, nb float64
	for i := range a {
		x := float64(a[i])
		y := float64(b[i])
		dot += x * y
		na += x * x
		nb += y * y
	}
	if na == 0 || nb == 0 {
		return 1
	}
	sim := dot / (math.Sqrt(na) * math.Sqrt(nb))
	return float32(1 - sim)
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
