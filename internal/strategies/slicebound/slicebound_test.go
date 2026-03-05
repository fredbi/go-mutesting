package slicebound_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/slicebound"
)

func discoverNodes(t *testing.T, src string) []mutation.Descriptor {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Uses: make(map[*ast.Ident]types.Object)}
	conf := types.Config{Error: func(err error) {}}
	pkg, _ := conf.Check("sample", fset, []*ast.File{file}, info)

	ctx := &strategy.DiscoveryContext{
		Fset: fset, File: file, Pkg: pkg, Info: info,
		Src: []byte(src), FilePath: "test.go", PkgPath: "sample",
	}

	targetTypes := map[string]bool{
		"*ast.IndexExpr": true,
		"*ast.SliceExpr": true,
	}

	var all []mutation.Descriptor
	for _, s := range strategy.All() {
		hasTarget := false
		for _, nt := range s.NodeTypes() {
			if targetTypes[nt] {
				hasTarget = true
				break
			}
		}
		if !hasTarget {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.(type) {
			case *ast.IndexExpr, *ast.SliceExpr:
				all = slices.AppendSeq(all, s.Discover(ctx, n))
			}
			return true
		})
	}
	return all
}

func byKind(descs []mutation.Descriptor, kind mutation.Kind) []mutation.Descriptor {
	var result []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == kind {
			result = append(result, d)
		}
	}
	return result
}

func TestIndexPlusOneMinus(t *testing.T) {
	src := `package sample

func First(items []int) int {
	return items[0]
}
`
	descs := discoverNodes(t, src)
	up := byKind(descs, mutation.SliceBoundIndexUp)
	down := byKind(descs, mutation.SliceBoundIndexDown)

	if len(up) != 1 {
		t.Fatalf("expected 1 index_plus_one, got %d", len(up))
	}
	if len(down) != 1 {
		t.Fatalf("expected 1 index_minus_one, got %d", len(down))
	}

	t.Logf("index+1: %s -> %s", up[0].Original, up[0].Replacement)
	t.Logf("index-1: %s -> %s", down[0].Original, down[0].Replacement)
}

func TestIndexVariable(t *testing.T) {
	src := `package sample

func At(items []int, i int) int {
	return items[i]
}
`
	descs := discoverNodes(t, src)
	up := byKind(descs, mutation.SliceBoundIndexUp)
	down := byKind(descs, mutation.SliceBoundIndexDown)

	if len(up) != 1 || len(down) != 1 {
		t.Fatalf("expected 1 up + 1 down, got %d + %d", len(up), len(down))
	}

	if up[0].Replacement != "i + 1" {
		t.Errorf("expected 'i + 1', got %q", up[0].Replacement)
	}
	if down[0].Replacement != "i - 1" {
		t.Errorf("expected 'i - 1', got %q", down[0].Replacement)
	}
}

func TestSliceHighPlusOne(t *testing.T) {
	src := `package sample

func Take(items []int, n int) []int {
	return items[0:n]
}
`
	descs := discoverNodes(t, src)
	hiUp := byKind(descs, mutation.SliceBoundHighUp)
	loUp := byKind(descs, mutation.SliceBoundLowUp)

	if len(hiUp) != 1 {
		t.Fatalf("expected 1 slice_high_plus_one, got %d", len(hiUp))
	}
	if hiUp[0].Replacement != "n + 1" {
		t.Errorf("expected 'n + 1', got %q", hiUp[0].Replacement)
	}

	if len(loUp) != 1 {
		t.Fatalf("expected 1 slice_low_plus_one, got %d", len(loUp))
	}
	if loUp[0].Replacement != "0 + 1" {
		t.Errorf("expected '0 + 1', got %q", loUp[0].Replacement)
	}
}

func TestSliceNoLow(t *testing.T) {
	src := `package sample

func Head(items []int, n int) []int {
	return items[:n]
}
`
	descs := discoverNodes(t, src)
	hiUp := byKind(descs, mutation.SliceBoundHighUp)
	loUp := byKind(descs, mutation.SliceBoundLowUp)

	if len(hiUp) != 1 {
		t.Fatalf("expected 1 slice_high_plus_one, got %d", len(hiUp))
	}
	// No low bound → no low mutation.
	if len(loUp) != 0 {
		t.Errorf("expected 0 slice_low_plus_one for [:n], got %d", len(loUp))
	}
}

func TestSliceNoHigh(t *testing.T) {
	src := `package sample

func Tail(items []int, n int) []int {
	return items[n:]
}
`
	descs := discoverNodes(t, src)
	hiUp := byKind(descs, mutation.SliceBoundHighUp)
	loUp := byKind(descs, mutation.SliceBoundLowUp)

	// No high bound → no high mutation.
	if len(hiUp) != 0 {
		t.Errorf("expected 0 slice_high_plus_one for [n:], got %d", len(hiUp))
	}
	if len(loUp) != 1 {
		t.Fatalf("expected 1 slice_low_plus_one, got %d", len(loUp))
	}
}
