package mutation

import (
	"crypto/sha256"
	"fmt"
)

// ComputeID returns the stable ID for a descriptor based on its
// file, kind, line, column, original, and replacement values.
//
// Uses line+column (not byte offset) for resilience to unrelated edits.
// Truncated to 32 hex chars.
func ComputeID(file string, kind Kind, line, col int, original, replacement string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\n%s\n%d\n%d\n%s\n%s", file, kind, line, col, original, replacement)
	return fmt.Sprintf("%x", h.Sum(nil))[:32]
}
