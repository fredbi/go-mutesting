package applier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

const structuralSrc = `package sample

func Compare(a, b int) bool {
	if a < b {
		return true
	} else {
		return false
	}
}

func Logic(a, b bool) bool {
	return a && b
}
`

func setupStructuralTest(t *testing.T) (string, string) {
	t.Helper()
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(structuralSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	return workdir, filePath
}

func TestStructuralApplierEmptyIfBlock(t *testing.T) {
	workdir, filePath := setupStructuralTest(t)

	// The if body starts at the '{' after "a < b".
	lbraceOffset := strings.Index(structuralSrc, "if a < b {\n\t\treturn true\n\t}")
	lbraceOffset = lbraceOffset + len("if a < b ")

	desc := mutation.Descriptor{
		ID:   "test-empty-if",
		File: filePath,
		Kind: mutation.BranchEmptyIf,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IfStmt",
				Action:      mutation.ActionEmptyBlock,
				TargetIndex: -1,
				StartOffset: lbraceOffset,
				EndOffset:   lbraceOffset + 20,
			},
		},
	}

	a := &StructuralApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	mutated, err := os.ReadFile(filepath.Join(workdir, filePath))
	if err != nil {
		t.Fatal(err)
	}

	// The if body should no longer contain "return true".
	if strings.Contains(string(mutated), "return true") {
		t.Errorf("expected if body to be emptied, got:\n%s", mutated)
	}
}

func TestStructuralApplierRollback(t *testing.T) {
	workdir, filePath := setupStructuralTest(t)

	lbraceOffset := strings.Index(structuralSrc, "if a < b {\n\t\treturn true\n\t}")
	lbraceOffset = lbraceOffset + len("if a < b ")

	desc := mutation.Descriptor{
		ID:   "test-rollback",
		File: filePath,
		Kind: mutation.BranchEmptyIf,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IfStmt",
				Action:      mutation.ActionEmptyBlock,
				TargetIndex: -1,
				StartOffset: lbraceOffset,
				EndOffset:   lbraceOffset + 20,
			},
		},
	}

	a := &StructuralApplier{}
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

	if !strings.Contains(string(restored), "return true") {
		t.Errorf("expected restored file to contain 'return true', got:\n%s", restored)
	}
}

func TestStructuralApplierReplaceExpr(t *testing.T) {
	workdir, filePath := setupStructuralTest(t)

	// Find "a" in "a && b" of the Logic function.
	logicIdx := strings.Index(structuralSrc, "return a && b")
	aOffset := logicIdx + len("return ")

	desc := mutation.Descriptor{
		ID:   "test-replace-expr",
		File: filePath,
		Kind: mutation.ExpressionRemoveTerm,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "BinaryExpr.X",
				Action:      mutation.ActionReplaceWithTrue,
				TargetIndex: -1,
				StartOffset: aOffset,
				EndOffset:   aOffset + 1,
			},
		},
	}

	a := &StructuralApplier{}
	if err := a.Apply(desc, workdir); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	mutated, err := os.ReadFile(filepath.Join(workdir, filePath))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(mutated), "true && b") {
		t.Errorf("expected 'true && b', got:\n%s", mutated)
	}
}
