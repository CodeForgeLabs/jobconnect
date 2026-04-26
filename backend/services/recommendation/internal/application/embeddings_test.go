package application

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type fakeEmbedder struct {
	mu       sync.Mutex
	calls    [][]string
	response [][]float32
	err      error
}

func (f *fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, append([]string(nil), texts...))
	if f.err != nil {
		return nil, f.err
	}
	if f.response != nil {
		return f.response, nil
	}
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = []float32{float32(i + 1), float32(i + 2)}
	}
	return out, nil
}

type fakeEmbeddingStore struct {
	mu      sync.Mutex
	entries map[string]StoredEmbedding
	getHits int
	upserts int
}

func newFakeEmbeddingStore() *fakeEmbeddingStore {
	return &fakeEmbeddingStore{entries: make(map[string]StoredEmbedding)}
}

func (f *fakeEmbeddingStore) Get(_ context.Context, st EmbeddingSourceType, id string) (StoredEmbedding, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getHits++
	e, ok := f.entries[string(st)+":"+id]
	return e, ok, nil
}

func (f *fakeEmbeddingStore) Upsert(_ context.Context, e StoredEmbedding) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.upserts++
	f.entries[string(e.SourceType)+":"+e.SourceID] = e
	return nil
}

func newResolveTestService(embedder Embedder, store EmbeddingStore) *RecommendationService {
	return NewRecommendationService(&fakeJobClient{}, &fakeUserClient{}, nil, nil, nil, embedder, store, newFreelancerTestConfig())
}

func TestResolveEmbeddingsShortCircuitsOnNoopEmbedder(t *testing.T) {
	svc := NewRecommendationService(&fakeJobClient{}, &fakeUserClient{}, nil, nil, nil, nil, newFakeEmbeddingStore(), newFreelancerTestConfig())
	got := svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeJob, SourceID: "1", Text: "hello"},
	})
	if got != nil {
		t.Fatalf("expected nil map when embedder is noop, got %v", got)
	}
}

func TestResolveEmbeddingsEmptyInput(t *testing.T) {
	svc := newResolveTestService(&fakeEmbedder{}, newFakeEmbeddingStore())
	if got := svc.resolveEmbeddings(context.Background(), nil); got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestResolveEmbeddingsComputesAndUpsertsOnMiss(t *testing.T) {
	emb := &fakeEmbedder{}
	store := newFakeEmbeddingStore()
	svc := newResolveTestService(emb, store)

	reqs := []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeFreelancer, SourceID: "u1", Text: "Go backend"},
		{SourceType: EmbeddingSourceTypeJob, SourceID: "42", Text: "Need Go engineer"},
	}
	got := svc.resolveEmbeddings(context.Background(), reqs)

	if len(got) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(got))
	}
	if _, ok := got.lookup(EmbeddingSourceTypeFreelancer, "u1"); !ok {
		t.Fatal("missing freelancer vector")
	}
	if _, ok := got.lookup(EmbeddingSourceTypeJob, "42"); !ok {
		t.Fatal("missing job vector")
	}
	if len(emb.calls) != 1 || len(emb.calls[0]) != 2 {
		t.Fatalf("expected one batched embed call with 2 texts, got %v", emb.calls)
	}
	if store.upserts != 2 {
		t.Fatalf("expected 2 upserts, got %d", store.upserts)
	}
}

func TestResolveEmbeddingsReusesStoreOnHashMatch(t *testing.T) {
	emb := &fakeEmbedder{}
	store := newFakeEmbeddingStore()

	// Pre-seed the store with the hash for our test text.
	text := "Go backend engineer"
	_ = store.Upsert(context.Background(), StoredEmbedding{
		SourceType: EmbeddingSourceTypeFreelancer,
		SourceID:   "u1",
		TextHash:   TextHash(text),
		Vector:     []float32{9, 9, 9},
	})

	svc := newResolveTestService(emb, store)
	got := svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeFreelancer, SourceID: "u1", Text: text},
	})

	vec, ok := got.lookup(EmbeddingSourceTypeFreelancer, "u1")
	if !ok {
		t.Fatal("expected cached vector lookup hit")
	}
	if len(vec) != 3 || vec[0] != 9 {
		t.Fatalf("expected seeded vector reused, got %v", vec)
	}
	if len(emb.calls) != 0 {
		t.Fatalf("embedder should not have been called, got %v", emb.calls)
	}
	if store.upserts != 1 {
		// 1 upsert from the seeding, 0 from the call
		t.Fatalf("expected no new upserts, got %d total", store.upserts)
	}
}

func TestResolveEmbeddingsRecomputesOnHashMismatch(t *testing.T) {
	emb := &fakeEmbedder{}
	store := newFakeEmbeddingStore()

	_ = store.Upsert(context.Background(), StoredEmbedding{
		SourceType: EmbeddingSourceTypeJob,
		SourceID:   "42",
		TextHash:   "stale-hash",
		Vector:     []float32{0, 0, 0},
	})

	svc := newResolveTestService(emb, store)
	text := "Updated job description"
	got := svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeJob, SourceID: "42", Text: text},
	})

	vec, ok := got.lookup(EmbeddingSourceTypeJob, "42")
	if !ok {
		t.Fatal("expected recomputed vector")
	}
	if vec[0] == 0 {
		t.Fatalf("expected fresh vector from embedder, got seeded zeros: %v", vec)
	}
	if len(emb.calls) != 1 {
		t.Fatalf("expected single embed call on hash mismatch, got %v", emb.calls)
	}
	after, _, _ := store.Get(context.Background(), EmbeddingSourceTypeJob, "42")
	if after.TextHash != TextHash(text) {
		t.Fatalf("store not updated with new hash: got %q want %q", after.TextHash, TextHash(text))
	}
}

func TestResolveEmbeddingsBatchesOnlyMisses(t *testing.T) {
	emb := &fakeEmbedder{}
	store := newFakeEmbeddingStore()

	// Seed one hit, leave others as misses.
	hitText := "already embedded"
	_ = store.Upsert(context.Background(), StoredEmbedding{
		SourceType: EmbeddingSourceTypeJob,
		SourceID:   "1",
		TextHash:   TextHash(hitText),
		Vector:     []float32{5, 5},
	})

	svc := newResolveTestService(emb, store)
	_ = svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeJob, SourceID: "1", Text: hitText},
		{SourceType: EmbeddingSourceTypeJob, SourceID: "2", Text: "miss-a"},
		{SourceType: EmbeddingSourceTypeJob, SourceID: "3", Text: "miss-b"},
	})

	if len(emb.calls) != 1 {
		t.Fatalf("expected a single batched embed call, got %d", len(emb.calls))
	}
	if len(emb.calls[0]) != 2 {
		t.Fatalf("expected batch of 2 miss texts, got %d", len(emb.calls[0]))
	}
}

func TestResolveEmbeddingsFallsBackWhenEmbedderFails(t *testing.T) {
	emb := &fakeEmbedder{err: ErrEmbedderUnavailable}
	store := newFakeEmbeddingStore()
	svc := newResolveTestService(emb, store)

	got := svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeJob, SourceID: "1", Text: "foo"},
	})
	if got != nil {
		t.Fatalf("expected nil fallback on embedder error, got %v", got)
	}
	if store.upserts != 0 {
		t.Fatalf("embedder error must not trigger upserts, got %d", store.upserts)
	}
}

func TestResolveEmbeddingsFallsBackOnShapeMismatch(t *testing.T) {
	emb := &fakeEmbedder{response: [][]float32{{1, 2}}} // returns 1 vec for 2 texts
	store := newFakeEmbeddingStore()
	svc := newResolveTestService(emb, store)

	got := svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeJob, SourceID: "1", Text: "a"},
		{SourceType: EmbeddingSourceTypeJob, SourceID: "2", Text: "b"},
	})
	if got != nil {
		t.Fatalf("expected nil fallback on shape mismatch, got %v", got)
	}
	if store.upserts != 0 {
		t.Fatalf("shape mismatch must not upsert, got %d", store.upserts)
	}
}

func TestResolveEmbeddingsPropagatesNonSentinelEmbedderError(t *testing.T) {
	emb := &fakeEmbedder{err: errors.New("boom")}
	store := newFakeEmbeddingStore()
	svc := newResolveTestService(emb, store)

	got := svc.resolveEmbeddings(context.Background(), []EmbeddingRequest{
		{SourceType: EmbeddingSourceTypeFreelancer, SourceID: "u1", Text: "x"},
	})
	if got != nil {
		t.Fatalf("expected nil fallback on generic embedder error, got %v", got)
	}
}
