package applier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

const returnSrc = `package sample

import "fmt"

func Validate(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return len(s), nil
}

func IsValid(s string) bool {
	return len(s) > 0
}
`

func TestStructuralApplierReturnZero(t *testing.T) {
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(returnSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	// Find the offset of "return 0, fmt.Errorf("empty")".
	retOffset := strings.Index(returnSrc, `return 0, fmt.Errorf("empty")`)

	desc := mutation.Descriptor{
		ID:   "test-return-nil-error",
		File: filePath,
		Kind: mutation.ReturnNilError,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "ReturnStmt",
				Action:      mutation.ActionReturnZero,
				TargetIndex: 1,
				StartOffset: retOffset,
				EndOffset:   retOffset + 30,
				ReturnMeta:  []string{"", "nil"},
			},
		},
	}

	a := &StructuralApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	mutated, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	content := string(mutated)
	if !strings.Contains(content, "return 0, nil") {
		t.Errorf("expected 'return 0, nil' in mutated file, got:\n%s", content)
	}
}

func TestStructuralApplierNegateBool(t *testing.T) {
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(returnSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	// Find "return len(s) > 0".
	retOffset := strings.Index(returnSrc, "return len(s) > 0")

	desc := mutation.Descriptor{
		ID:   "test-negate-bool",
		File: filePath,
		Kind: mutation.ReturnNegateBool,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "ReturnStmt",
				Action:      mutation.ActionNegateBoolReturn,
				TargetIndex: 0,
				StartOffset: retOffset,
				EndOffset:   retOffset + 18,
				ReturnMeta:  []string{"!"},
			},
		},
	}

	a := &StructuralApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	mutated, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	content := string(mutated)
	// The boolean return should be negated.
	if !strings.Contains(content, "!(len(s) > 0)") && !strings.Contains(content, "!len(s) > 0") {
		t.Errorf("expected negated bool return, got:\n%s", content)
	}
}

func TestStructuralApplierReturnZeroRollback(t *testing.T) {
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(returnSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	retOffset := strings.Index(returnSrc, `return 0, fmt.Errorf("empty")`)

	desc := mutation.Descriptor{
		ID:   "test-return-rb",
		File: filePath,
		Kind: mutation.ReturnNilError,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "ReturnStmt",
				Action:      mutation.ActionReturnZero,
				TargetIndex: 1,
				StartOffset: retOffset,
				EndOffset:   retOffset + 30,
				ReturnMeta:  []string{"", "nil"},
			},
		},
	}

	a := &StructuralApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatal(err)
	}
	if err := a.Rollback(desc, workdir); err != nil {
		t.Fatal(err)
	}

	restored, _ := os.ReadFile(target)
	if !strings.Contains(string(restored), `fmt.Errorf("empty")`) {
		t.Errorf("expected original error expression after rollback, got:\n%s", restored)
	}
}
