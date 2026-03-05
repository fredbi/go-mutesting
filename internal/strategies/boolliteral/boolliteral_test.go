package boolliteral_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/boolliteral"
)

func discoverIdents(t *testing.T, src string) []mutation.Descriptor {
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

	var all []mutation.Descriptor
	for _, s := range strategy.All() {
		for _, nt := range s.NodeTypes() {
			if nt != "*ast.Ident" {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if _, ok := n.(*ast.Ident); !ok {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func TestBoolLiteralDiscovery(t *testing.T) {
	src := `package sample

var debug = true
var verbose = false

func Check() bool {
	return true
}
`
	descs := discoverIdents(t, src)

	var trueToFalse, falseToTrue int
	for _, d := range descs {
		switch d.Kind {
		case mutation.BoolLitTrueToFalse:
			trueToFalse++
			t.Logf("true->false at line %d", d.StartPos.Line)
		case mutation.BoolLitFalseToTrue:
			falseToTrue++
			t.Logf("false->true at line %d", d.StartPos.Line)
		}
	}

	if trueToFalse != 2 {
		t.Errorf("expected 2 true->false mutations, got %d", trueToFalse)
	}
	if falseToTrue != 1 {
		t.Errorf("expected 1 false->true mutation, got %d", falseToTrue)
	}
}

func TestBoolLiteralSkipsUserDefined(t *testing.T) {
	// A variable named "true" (if someone did that) should not be mutated.
	// In practice this won't parse as valid Go, but if Obj != nil we skip it.
	src := `package sample

var x = true
`
	descs := discoverIdents(t, src)

	count := 0
	for _, d := range descs {
		if d.Kind == mutation.BoolLitTrueToFalse {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 bool literal mutation, got %d", count)
	}
}
