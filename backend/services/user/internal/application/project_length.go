package application

import "strings"

const (
	ProjectLengthUnspecified = "PROJECT_LENGTH_UNSPECIFIED"
	ProjectLengthShortTerm   = "PROJECT_LENGTH_SHORT_TERM"
	ProjectLengthMediumTerm  = "PROJECT_LENGTH_MEDIUM_TERM"
	ProjectLengthLongTerm    = "PROJECT_LENGTH_LONG_TERM"
)

// CanonicalProjectLength converts supported project length inputs to canonical tokens.
func CanonicalProjectLength(value string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case "", "UNSPECIFIED", "PROJECT_LENGTH_UNSPECIFIED", "NO_PREFERENCE", "NONE":
		return ProjectLengthUnspecified, true
	case "SHORT", "SHORT_TERM", "SHORT-TERM", "PROJECT_LENGTH_SHORT_TERM":
		return ProjectLengthShortTerm, true
	case "MEDIUM", "MEDIUM_TERM", "MEDIUM-TERM", "PROJECT_LENGTH_MEDIUM_TERM":
		return ProjectLengthMediumTerm, true
	case "LONG", "LONG_TERM", "LONG-TERM", "PROJECT_LENGTH_LONG_TERM":
		return ProjectLengthLongTerm, true
	default:
		return "", false
	}
}

func CanonicalProjectLengthOrUnspecified(value string) string {
	canonical, ok := CanonicalProjectLength(value)
	if !ok {
		return ProjectLengthUnspecified
	}
	return canonical
}

func IsProjectLengthPreferenceSet(value string) bool {
	canonical, ok := CanonicalProjectLength(value)
	if !ok {
		return false
	}
	return canonical != ProjectLengthUnspecified
}
