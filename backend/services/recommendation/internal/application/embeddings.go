package application

import (
	"context"
	"errors"
	"log"
)

// EmbeddingRequest identifies a single text whose embedding the ranker wants.
// SourceID is used as the map key in the result returned by resolveEmbeddings,
// so callers must ensure uniqueness within a single call (in practice a
// request mixes at most one entity per source type).
type EmbeddingRequest struct {
	SourceType EmbeddingSourceType
	SourceID   string
	Text       string
}

// embeddingVectors maps (sourceType, sourceID) to the vector for that source.
type embeddingVectors map[string][]float32

func (v embeddingVectors) lookup(sourceType EmbeddingSourceType, sourceID string) ([]float32, bool) {
	if v == nil {
		return nil, false
	}
	vec, ok := v[embeddingMapKey(sourceType, sourceID)]
	return vec, ok
}

func embeddingMapKey(sourceType EmbeddingSourceType, sourceID string) string {
	return string(sourceType) + ":" + sourceID
}

// resolveEmbeddings returns a vector for each request whose text the embedder
// can (or already did) encode. It short-circuits when the embedder is a noop.
// The semantics are:
//
//   - Store hit with matching text hash → reuse the stored vector.
//   - Store miss or hash mismatch → queue the text for a batched Embed call.
//   - Single batched Embedder.Embed is issued for all misses.
//   - Successful batch vectors are upserted back into the store.
//   - If the Embedder call fails for any reason — including
//     ErrEmbedderUnavailable — the ranker must not be blocked: this returns
//     a nil map and the caller falls back to token cosine for that request.
func (s *RecommendationService) resolveEmbeddings(ctx context.Context, reqs []EmbeddingRequest) embeddingVectors {
	if len(reqs) == 0 {
		return nil
	}
	if _, noop := s.embedder.(noopEmbedder); noop {
		return nil
	}

	out := make(embeddingVectors, len(reqs))
	var missIdx []int
	var missTexts []string

	for i, req := range reqs {
		hash := TextHash(req.Text)
		existing, found, err := s.embeddingStore.Get(ctx, req.SourceType, req.SourceID)
		if err == nil && found && existing.TextHash == hash && len(existing.Vector) > 0 {
			out[embeddingMapKey(req.SourceType, req.SourceID)] = existing.Vector
			continue
		}
		missIdx = append(missIdx, i)
		missTexts = append(missTexts, req.Text)
	}

	if len(missTexts) == 0 {
		return out
	}

	vectors, err := s.embedder.Embed(ctx, missTexts)
	if err != nil {
		if !errors.Is(err, ErrEmbedderUnavailable) {
			log.Printf("recommendation: embedder call failed, falling back to token cosine: %v", err)
		}
		return nil
	}
	if len(vectors) != len(missTexts) {
		log.Printf("recommendation: embedder shape mismatch want=%d got=%d, falling back to token cosine", len(missTexts), len(vectors))
		return nil
	}

	for j, vec := range vectors {
		req := reqs[missIdx[j]]
		out[embeddingMapKey(req.SourceType, req.SourceID)] = vec
		if err := s.embeddingStore.Upsert(ctx, StoredEmbedding{
			SourceType: req.SourceType,
			SourceID:   req.SourceID,
			TextHash:   TextHash(req.Text),
			Vector:     vec,
		}); err != nil {
			log.Printf("recommendation: embedding upsert failed source=%s/%s: %v", req.SourceType, req.SourceID, err)
		}
	}

	return out
}
