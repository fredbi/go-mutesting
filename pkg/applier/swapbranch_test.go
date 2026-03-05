package applier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

const swapIfElseSrc = `package sample

func Compare(a, b int) bool {
	if a < b {
		return true
	} else {
		return false
	}
}
`

func TestStructuralApplierSwapIfElse(t *testing.T) {
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(swapIfElseSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	// Find the offset of the if-body opening brace.
	lbraceOffset := strings.Index(swapIfElseSrc, "if a < b {") + len("if a < b ")

	desc := mutation.Descriptor{
		ID:   "test-swap-if-else",
		File: filePath,
		Kind: mutation.BranchSwapIfElse,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IfStmt.SwapElse",
				Action:      mutation.ActionSwapIfElse,
				TargetIndex: -1,
				StartOffset: lbraceOffset,
				EndOffset:   lbraceOffset + 50,
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
	// After swap: if body should have "return false" and else body "return true".
	ifIdx := strings.Index(content, "if a < b {")
	elseIdx := strings.Index(content, "} else {")
	if ifIdx < 0 || elseIdx < 0 {
		t.Fatalf("couldn't find if/else structure in mutated file:\n%s", content)
	}

	ifBody := content[ifIdx:elseIdx]
	elseBody := content[elseIdx:]

	if !strings.Contains(ifBody, "return false") {
		t.Errorf("expected if body to contain 'return false' after swap, got:\n%s", ifBody)
	}
	if !strings.Contains(elseBody, "return true") {
		t.Errorf("expected else body to contain 'return true' after swap, got:\n%s", elseBody)
	}
}

const swapCaseSrc = `package sample

func SwitchCase(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	default:
		return "other"
	}
}
`

func TestStructuralApplierSwapCase(t *testing.T) {
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(swapCaseSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	// Find the offset of the first case colon.
	colonOffset := strings.Index(swapCaseSrc, "case 1:") + len("case 1")

	desc := mutation.Descriptor{
		ID:   "test-swap-case",
		File: filePath,
		Kind: mutation.BranchSwapCase,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:     "SwitchStmt.SwapCase",
				Action:       mutation.ActionSwapCase,
				TargetIndex:  0,
				TargetIndex2: 1,
				StartOffset:  colonOffset,
				EndOffset:    colonOffset + 50,
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
	// After swap: case 1 should return "two", case 2 should return "one".
	case1Idx := strings.Index(content, "case 1:")
	case2Idx := strings.Index(content, "case 2:")
	if case1Idx < 0 || case2Idx < 0 {
		t.Fatalf("couldn't find case structure in mutated file:\n%s", content)
	}

	case1Body := content[case1Idx:case2Idx]
	if !strings.Contains(case1Body, `"two"`) {
		t.Errorf("expected case 1 body to contain '\"two\"' after swap, got:\n%s", case1Body)
	}
}

func TestStructuralApplierSwapIfElseRollback(t *testing.T) {
	workdir := t.TempDir()
	filePath := "test.go"
	target := filepath.Join(workdir, filePath)
	if err := os.WriteFile(target, []byte(swapIfElseSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	lbraceOffset := strings.Index(swapIfElseSrc, "if a < b {") + len("if a < b ")

	desc := mutation.Descriptor{
		ID:   "test-swap-rb",
		File: filePath,
		Kind: mutation.BranchSwapIfElse,
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IfStmt.SwapElse",
				Action:      mutation.ActionSwapIfElse,
				TargetIndex: -1,
				StartOffset: lbraceOffset,
				EndOffset:   lbraceOffset + 50,
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
	// After rollback the if body should have "return true" again.
	ifIdx := strings.Index(string(restored), "if a < b {")
	elseIdx := strings.Index(string(restored), "} else {")
	ifBody := string(restored)[ifIdx:elseIdx]
	if !strings.Contains(ifBody, "return true") {
		t.Errorf("expected 'return true' in if body after rollback, got:\n%s", ifBody)
	}
}
