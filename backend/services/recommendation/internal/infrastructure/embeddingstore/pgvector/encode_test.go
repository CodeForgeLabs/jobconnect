package pgvector

import (
	"strings"
	"testing"
)

func TestEncodeVectorFormat(t *testing.T) {
	got := encodeVector([]float32{1, 2.5, -0.5})
	want := "[1,2.5,-0.5]"
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestEncodeVectorEmpty(t *testing.T) {
	if got := encodeVector(nil); got != "[]" {
		t.Fatalf("got=%q want=%q", got, "[]")
	}
}

func TestDecodeVectorRoundTrip(t *testing.T) {
	want := []float32{1, 2.5, -0.5, 0}
	got, err := decodeVector(encodeVector(want))
	if err != nil {
		t.Fatalf("decodeVector: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len got=%d want=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("[%d] got=%v want=%v", i, got[i], want[i])
		}
	}
}

func TestDecodeVectorEmptyInner(t *testing.T) {
	got, err := decodeVector("[]")
	if err != nil {
		t.Fatalf("decodeVector: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestDecodeVectorRejectsMalformed(t *testing.T) {
	cases := []string{"", "1,2,3", "[1,2", "1,2]", "[a,b]"}
	for _, in := range cases {
		if _, err := decodeVector(in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestResolveTargetFreelancer(t *testing.T) {
	table, col, val, err := resolveTarget("freelancer", "user-123")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if table != "freelancer_embeddings" || col != "user_id" {
		t.Fatalf("table=%q col=%q", table, col)
	}
	if s, ok := val.(string); !ok || s != "user-123" {
		t.Fatalf("val=%v ok-string=%v", val, ok)
	}
}

func TestResolveTargetJob(t *testing.T) {
	_, col, val, err := resolveTarget("job", "42")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if col != "job_id" {
		t.Fatalf("col=%q", col)
	}
	if id, ok := val.(int64); !ok || id != 42 {
		t.Fatalf("val=%v ok-int64=%v", val, ok)
	}
}

func TestResolveTargetRejectsBadJobID(t *testing.T) {
	_, _, _, err := resolveTarget("job", "not-a-number")
	if err == nil || !strings.Contains(err.Error(), "int64") {
		t.Fatalf("expected int64 parse error, got %v", err)
	}
}

func TestResolveTargetRejectsEmptyFreelancerID(t *testing.T) {
	if _, _, _, err := resolveTarget("freelancer", "  "); err == nil {
		t.Fatal("expected error for empty freelancer id")
	}
}

func TestResolveTargetRejectsUnknownType(t *testing.T) {
	if _, _, _, err := resolveTarget("mystery", "x"); err == nil {
		t.Fatal("expected error for unknown source type")
	}
}
