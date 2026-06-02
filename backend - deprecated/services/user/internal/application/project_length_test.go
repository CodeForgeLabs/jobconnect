package application

import "testing"

func TestCanonicalProjectLength(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
		wantOK    bool
	}{
		{name: "empty maps to unspecified", input: "", wantValue: ProjectLengthUnspecified, wantOK: true},
		{name: "legacy short alias", input: "short", wantValue: ProjectLengthShortTerm, wantOK: true},
		{name: "enum medium token", input: "PROJECT_LENGTH_MEDIUM_TERM", wantValue: ProjectLengthMediumTerm, wantOK: true},
		{name: "hyphenated long alias", input: "long-term", wantValue: ProjectLengthLongTerm, wantOK: true},
		{name: "invalid value", input: "seasonal", wantValue: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOK := CanonicalProjectLength(tt.input)
			if gotOK != tt.wantOK {
				t.Fatalf("ok mismatch: got %v, want %v", gotOK, tt.wantOK)
			}
			if gotValue != tt.wantValue {
				t.Fatalf("value mismatch: got %q, want %q", gotValue, tt.wantValue)
			}
		})
	}
}

func TestIsProjectLengthPreferenceSet(t *testing.T) {
	if IsProjectLengthPreferenceSet("") {
		t.Fatalf("expected no preference for empty value")
	}
	if IsProjectLengthPreferenceSet(ProjectLengthUnspecified) {
		t.Fatalf("expected unspecified to not count as preference")
	}
	if !IsProjectLengthPreferenceSet("short") {
		t.Fatalf("expected short to count as preference")
	}
}
