package application

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

// TextHash returns a 16-character hex digest of the normalised text. The
// normalisation is intentionally coarse — trim + collapse whitespace +
// lowercase — so trivial whitespace or casing edits do not trigger re-embed,
// but any substantive content change does.
func TextHash(text string) string {
	sum := sha256.Sum256([]byte(normalizeForHash(text)))
	return hex.EncodeToString(sum[:8])
}

func normalizeForHash(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true // skip leading whitespace
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(unicode.ToLower(r))
		prevSpace = false
	}
	out := b.String()
	return strings.TrimRight(out, " ")
}
