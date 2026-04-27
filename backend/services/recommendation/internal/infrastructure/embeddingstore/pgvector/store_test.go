package pgvector

import (
	"context"
	"os"
	"testing"
	"time"

	"jobconnect/recommendation/internal/application"
)

// requirePostgres skips the test unless RECOMMENDATION_TEST_POSTGRES_URL is
// set. The expected target is a Postgres database with the pgvector extension
// available and the recommendation 0001_init migration applied. To run:
//
//	docker compose up -d postgres
//	./scripts/migrate-all.sh
//	RECOMMENDATION_TEST_POSTGRES_URL='postgres://recommendation:recommendation@localhost:5432/jobconnect_recommendation?sslmode=disable' \
//	  go test ./internal/infrastructure/embeddingstore/pgvector/...
func requirePostgres(t *testing.T) string {
	t.Helper()
	url := os.Getenv("RECOMMENDATION_TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("RECOMMENDATION_TEST_POSTGRES_URL is not set; skipping pgvector integration test")
	}
	return url
}

func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	url := requirePostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanCancel()
	if _, err := pool.Exec(cleanCtx, "TRUNCATE freelancer_embeddings, job_embeddings"); err != nil {
		pool.Close()
		t.Fatalf("truncate: %v", err)
	}
	return New(pool), func() { pool.Close() }
}

func TestPgvectorStoreUpsertAndGetFreelancer(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	want := application.StoredEmbedding{
		SourceType: application.EmbeddingSourceTypeFreelancer,
		SourceID:   "freelancer-test-1",
		TextHash:   "hash-abc",
		Vector:     buildTestVector(384, 0.5),
	}
	if err := store.Upsert(ctx, want); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, ok, err := store.Get(ctx, want.SourceType, want.SourceID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected stored row to be found")
	}
	if got.TextHash != want.TextHash {
		t.Fatalf("TextHash got=%q want=%q", got.TextHash, want.TextHash)
	}
	if len(got.Vector) != len(want.Vector) {
		t.Fatalf("Vector length got=%d want=%d", len(got.Vector), len(want.Vector))
	}
	if !approxEqual(got.Vector[0], want.Vector[0]) {
		t.Fatalf("Vector[0] got=%v want=%v", got.Vector[0], want.Vector[0])
	}
}

func TestPgvectorStoreUpsertAndGetJob(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	want := application.StoredEmbedding{
		SourceType: application.EmbeddingSourceTypeJob,
		SourceID:   "424242",
		TextHash:   "hash-job",
		Vector:     buildTestVector(384, -0.25),
	}
	if err := store.Upsert(ctx, want); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, ok, err := store.Get(ctx, want.SourceType, want.SourceID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected stored job row to be found")
	}
	if got.TextHash != want.TextHash || got.SourceID != want.SourceID {
		t.Fatalf("mismatch: got=%+v want=%+v", got, want)
	}
}

func TestPgvectorStoreUpsertOverwritesOnConflict(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	first := application.StoredEmbedding{
		SourceType: application.EmbeddingSourceTypeFreelancer,
		SourceID:   "freelancer-test-overwrite",
		TextHash:   "hash-v1",
		Vector:     buildTestVector(384, 0.1),
	}
	second := first
	second.TextHash = "hash-v2"
	second.Vector = buildTestVector(384, 0.9)

	if err := store.Upsert(ctx, first); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}
	if err := store.Upsert(ctx, second); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	got, ok, err := store.Get(ctx, first.SourceType, first.SourceID)
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if got.TextHash != "hash-v2" {
		t.Fatalf("TextHash not overwritten: got=%q", got.TextHash)
	}
	if !approxEqual(got.Vector[0], second.Vector[0]) {
		t.Fatalf("Vector not overwritten: got[0]=%v want=%v", got.Vector[0], second.Vector[0])
	}
}

func TestPgvectorStoreGetReturnsFalseOnMiss(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	_, ok, err := store.Get(context.Background(), application.EmbeddingSourceTypeFreelancer, "missing-user")
	if err != nil {
		t.Fatalf("Get on miss returned error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false on miss")
	}
}

func TestPgvectorStoreSearchByVectorOrdersByDistance(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	ctx := context.Background()
	near := buildOrientedVector(384, 0)
	far := buildOrientedVector(384, 1)
	mid := buildBlendedVector(384, 0, 1)

	if err := store.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "1", TextHash: "h-near", Vector: near}); err != nil {
		t.Fatalf("Upsert near: %v", err)
	}
	if err := store.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "2", TextHash: "h-far", Vector: far}); err != nil {
		t.Fatalf("Upsert far: %v", err)
	}
	if err := store.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "3", TextHash: "h-mid", Vector: mid}); err != nil {
		t.Fatalf("Upsert mid: %v", err)
	}

	hits, err := store.SearchByVector(ctx, application.EmbeddingSourceTypeJob, near, 3)
	if err != nil {
		t.Fatalf("SearchByVector: %v", err)
	}
	if len(hits) != 3 {
		t.Fatalf("expected 3 hits, got %d", len(hits))
	}
	if hits[0].SourceID != "1" {
		t.Fatalf("expected nearest hit to be job 1, got %+v", hits)
	}
	for i := 1; i < len(hits); i++ {
		if hits[i].Distance < hits[i-1].Distance {
			t.Fatalf("hits not sorted ascending: %+v", hits)
		}
	}
}

func buildOrientedVector(dim, axis int) []float32 {
	v := make([]float32, dim)
	v[axis] = 1
	return v
}

func buildBlendedVector(dim, a, b int) []float32 {
	v := make([]float32, dim)
	v[a] = 0.7071
	v[b] = 0.7071
	return v
}

func buildTestVector(dim int, base float32) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = base + float32(i)*0.001
	}
	return v
}

func approxEqual(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-3
}
