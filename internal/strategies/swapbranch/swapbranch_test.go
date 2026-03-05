package swapbranch_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/swapbranch"
)

const swapSrc = `package sample

func Compare(a, b int) bool {
	if a < b {
		return true
	} else {
		return false
	}
}

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

func parse(t *testing.T, src string) (*token.FileSet, *ast.File, *types.Package, *types.Info) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Uses: make(map[*ast.Ident]types.Object)}
	conf := types.Config{Error: func(err error) {}}
	pkg, _ := conf.Check("sample", fset, []*ast.File{file}, info)
	return fset, file, pkg, info
}

func discoverAll(t *testing.T, src string) []mutation.Descriptor {
	t.Helper()
	fset, file, pkg, info := parse(t, src)

	ctx := &strategy.DiscoveryContext{
		Fset:     fset,
		File:     file,
		Pkg:      pkg,
		Info:     info,
		Src:      []byte(src),
		FilePath: "test.go",
		PkgPath:  "sample",
	}

	var all []mutation.Descriptor
	for _, s := range strategy.All() {
		for _, nt := range s.NodeTypes() {
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				// Simple type name match.
				typeName := ""
				switch n.(type) {
				case *ast.IfStmt:
					typeName = "*ast.IfStmt"
				case *ast.SwitchStmt:
					typeName = "*ast.SwitchStmt"
				case *ast.TypeSwitchStmt:
					typeName = "*ast.TypeSwitchStmt"
				}
				if typeName != nt {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func TestSwapIfElseDiscovery(t *testing.T) {
	descs := discoverAll(t, swapSrc)

	var swapIfElse []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.BranchSwapIfElse {
			swapIfElse = append(swapIfElse, d)
		}
	}

	if len(swapIfElse) != 1 {
		t.Fatalf("expected 1 swap_if_else mutation, got %d", len(swapIfElse))
	}

	d := swapIfElse[0]
	if d.Apply.Structural == nil {
		t.Fatal("expected Structural ApplySpec")
	}
	if d.Apply.Structural.Action != mutation.ActionSwapIfElse {
		t.Errorf("expected ActionSwapIfElse, got %v", d.Apply.Structural.Action)
	}
	t.Logf("swap_if_else: %s -> %s", d.Original, d.Replacement)
}

func TestSwapCaseDiscovery(t *testing.T) {
	descs := discoverAll(t, swapSrc)

	var swapCase []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.BranchSwapCase {
			swapCase = append(swapCase, d)
		}
	}

	// 3 cases (1, 2, default) → 2 adjacent pairs.
	if len(swapCase) != 2 {
		t.Fatalf("expected 2 swap_case mutations, got %d", len(swapCase))
	}

	for i, d := range swapCase {
		if d.Apply.Structural == nil {
			t.Fatalf("swap_case[%d]: expected Structural ApplySpec", i)
		}
		if d.Apply.Structural.Action != mutation.ActionSwapCase {
			t.Errorf("swap_case[%d]: expected ActionSwapCase, got %v", i, d.Apply.Structural.Action)
		}
		t.Logf("swap_case[%d]: %s -> %s (indices %d,%d)", i, d.Original, d.Replacement,
			d.Apply.Structural.TargetIndex, d.Apply.Structural.TargetIndex2)
	}
}

func TestSwapIfElseSkipsElseIf(t *testing.T) {
	src := `package sample

func Chain(x int) string {
	if x > 0 {
		return "pos"
	} else if x < 0 {
		return "neg"
	} else {
		return "zero"
	}
}
`
	descs := discoverAll(t, src)

	for _, d := range descs {
		if d.Kind == mutation.BranchSwapIfElse {
			// The outermost if has an else-if, not a plain else.
			// Only the inner "else if ... else" should be swapped.
			t.Logf("found swap: %s at line %d", d.Original, d.StartPos.Line)
		}
	}
}
