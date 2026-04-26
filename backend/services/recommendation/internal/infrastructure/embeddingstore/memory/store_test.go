package memory

import (
	"context"
	"testing"

	"jobconnect/recommendation/internal/application"
)

func TestGetMissReturnsFalse(t *testing.T) {
	s := New()
	_, ok, err := s.Get(context.Background(), application.EmbeddingSourceTypeJob, "42")
	if err != nil {
		t.Fatalf("Get err: %v", err)
	}
	if ok {
		t.Fatal("expected miss, got hit")
	}
}

func TestUpsertAndGetRoundTrip(t *testing.T) {
	s := New()
	e := application.StoredEmbedding{
		SourceType: application.EmbeddingSourceTypeJob,
		SourceID:   "42",
		TextHash:   "abc",
		Vector:     []float32{1, 2, 3},
	}
	if err := s.Upsert(context.Background(), e); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, ok, err := s.Get(context.Background(), application.EmbeddingSourceTypeJob, "42")
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if got.TextHash != "abc" || len(got.Vector) != 3 || got.Vector[2] != 3 {
		t.Fatalf("unexpected stored embedding: %+v", got)
	}
}

func TestUpsertReplacesExisting(t *testing.T) {
	s := New()
	ctx := context.Background()
	_ = s.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "42", TextHash: "h1", Vector: []float32{1}})
	_ = s.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "42", TextHash: "h2", Vector: []float32{2, 2}})

	got, _, _ := s.Get(ctx, application.EmbeddingSourceTypeJob, "42")
	if got.TextHash != "h2" || len(got.Vector) != 2 {
		t.Fatalf("upsert did not replace: %+v", got)
	}
	if s.Len() != 1 {
		t.Fatalf("Len = %d, want 1", s.Len())
	}
}

func TestSourceTypesAreIsolated(t *testing.T) {
	s := New()
	ctx := context.Background()
	_ = s.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "42", TextHash: "job"})
	_ = s.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeFreelancer, SourceID: "42", TextHash: "freelancer"})

	job, _, _ := s.Get(ctx, application.EmbeddingSourceTypeJob, "42")
	fl, _, _ := s.Get(ctx, application.EmbeddingSourceTypeFreelancer, "42")
	if job.TextHash != "job" || fl.TextHash != "freelancer" {
		t.Fatalf("source types collided: job=%+v freelancer=%+v", job, fl)
	}
}

func TestStoredVectorIsClonedOnRead(t *testing.T) {
	s := New()
	ctx := context.Background()
	orig := []float32{1, 2, 3}
	_ = s.Upsert(ctx, application.StoredEmbedding{SourceType: application.EmbeddingSourceTypeJob, SourceID: "1", Vector: orig})

	got, _, _ := s.Get(ctx, application.EmbeddingSourceTypeJob, "1")
	got.Vector[0] = 99

	again, _, _ := s.Get(ctx, application.EmbeddingSourceTypeJob, "1")
	if again.Vector[0] != 1 {
		t.Fatalf("caller mutation leaked into store: %v", again.Vector)
	}
}
