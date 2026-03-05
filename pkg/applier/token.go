package applier

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// TokenApplier applies token-level mutations by byte-splicing.
type TokenApplier struct{}

func (a *TokenApplier) Apply(desc mutation.Descriptor, workdir string) error {
	spec := desc.Apply.TokenSwap
	if spec == nil {
		return fmt.Errorf("descriptor %s: no TokenSwapSpec", desc.ID)
	}

	target := filepath.Join(workdir, desc.File)
	src, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("reading %s: %w", target, err)
	}

	if spec.StartOffset < 0 || spec.EndOffset > len(src) || spec.StartOffset > spec.EndOffset {
		return fmt.Errorf("descriptor %s: offset [%d, %d) out of range for file of %d bytes",
			desc.ID, spec.StartOffset, spec.EndOffset, len(src))
	}

	// Validate original token at recorded position.
	actual := string(src[spec.StartOffset:spec.EndOffset])
	if actual != spec.OriginalToken {
		return fmt.Errorf("descriptor %s: expected %q at offset %d, found %q",
			desc.ID, spec.OriginalToken, spec.StartOffset, actual)
	}

	// Byte-splice the replacement.
	var buf bytes.Buffer
	buf.Write(src[:spec.StartOffset])
	buf.WriteString(spec.ReplacementToken)
	buf.Write(src[spec.EndOffset:])

	mutated := buf.Bytes()

	// Validate result with go/format.
	formatted, err := format.Source(mutated)
	if err != nil {
		return fmt.Errorf("descriptor %s: mutated file does not parse: %w", desc.ID, err)
	}

	return os.WriteFile(target, formatted, 0o644)
}

func (a *TokenApplier) Rollback(desc mutation.Descriptor, workdir string) error {
	spec := desc.Apply.TokenSwap
	if spec == nil {
		return fmt.Errorf("descriptor %s: no TokenSwapSpec", desc.ID)
	}

	target := filepath.Join(workdir, desc.File)
	src, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("reading %s: %w", target, err)
	}

	// Find the replacement token and swap back.
	repBytes := []byte(spec.ReplacementToken)
	idx := bytes.Index(src[spec.StartOffset:], repBytes)
	if idx != 0 {
		return fmt.Errorf("descriptor %s: cannot find replacement token %q at offset %d for rollback",
			desc.ID, spec.ReplacementToken, spec.StartOffset)
	}

	var buf bytes.Buffer
	buf.Write(src[:spec.StartOffset])
	buf.WriteString(spec.OriginalToken)
	buf.Write(src[spec.StartOffset+len(repBytes):])

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("descriptor %s: rollback file does not parse: %w", desc.ID, err)
	}

	return os.WriteFile(target, formatted, 0o644)
}
