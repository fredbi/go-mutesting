package applier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

const sampleSrc = `package sample

func Add(a, b int) int {
	return a + b
}
`

func setupTokenTest(t *testing.T) (string, string) {
	t.Helper()
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(sampleSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	return workdir, filePath
}

func TestTokenApplierApply(t *testing.T) {
	workdir, filePath := setupTokenTest(t)

	// Find the offset of '+' in "a + b".
	offset := strings.Index(sampleSrc, "a + b") + 2 // points to '+'

	desc := mutation.Descriptor{
		ID:   "test-token",
		File: filePath,
		Kind: mutation.ArithmeticAddToSub,
		Apply: mutation.ApplySpec{
			TokenSwap: &mutation.TokenSwapSpec{
				OriginalToken:    "+",
				ReplacementToken: "-",
				StartOffset:      offset,
				EndOffset:        offset + 1,
			},
		},
	}

	a := &TokenApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	mutated, err := os.ReadFile(filepath.Join(workdir, filePath))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(mutated), "a - b") {
		t.Errorf("expected mutated file to contain 'a - b', got:\n%s", mutated)
	}
}

func TestTokenApplierRollback(t *testing.T) {
	workdir, filePath := setupTokenTest(t)

	offset := strings.Index(sampleSrc, "a + b") + 2

	desc := mutation.Descriptor{
		ID:   "test-token-rb",
		File: filePath,
		Kind: mutation.ArithmeticAddToSub,
		Apply: mutation.ApplySpec{
			TokenSwap: &mutation.TokenSwapSpec{
				OriginalToken:    "+",
				ReplacementToken: "-",
				StartOffset:      offset,
				EndOffset:        offset + 1,
			},
		},
	}

	a := &TokenApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if err := a.Rollback(desc, workdir); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	restored, err := os.ReadFile(filepath.Join(workdir, filePath))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(restored), "a + b") {
		t.Errorf("expected restored file to contain 'a + b', got:\n%s", restored)
	}
}

func TestTokenApplierBadOffset(t *testing.T) {
	workdir, filePath := setupTokenTest(t)

	desc := mutation.Descriptor{
		ID:   "test-bad-offset",
		File: filePath,
		Apply: mutation.ApplySpec{
			TokenSwap: &mutation.TokenSwapSpec{
				OriginalToken:    "+",
				ReplacementToken: "-",
				StartOffset:      9999,
				EndOffset:        10000,
			},
		},
	}

	a := &TokenApplier{}
	if err := a.Apply(desc, workdir); err == nil {
		t.Error("expected error for out-of-range offset")
	}
}

func TestTokenApplierMismatch(t *testing.T) {
	workdir, filePath := setupTokenTest(t)

	desc := mutation.Descriptor{
		ID:   "test-mismatch",
		File: filePath,
		Apply: mutation.ApplySpec{
			TokenSwap: &mutation.TokenSwapSpec{
				OriginalToken:    "*",
				ReplacementToken: "/",
				StartOffset:      0,
				EndOffset:        1,
			},
		},
	}

	a := &TokenApplier{}
	if err := a.Apply(desc, workdir); err == nil {
		t.Error("expected error for token mismatch")
	}
}
