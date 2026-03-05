package stringliteral_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/stringliteral"
)

func discoverBasicLits(t *testing.T, src string) []mutation.Descriptor {
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
			if nt != "*ast.BasicLit" {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if _, ok := n.(*ast.BasicLit); !ok {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func TestStringLiteralNonEmptyToEmpty(t *testing.T) {
	src := `package sample

var greeting = "hello"
var name = "world"
`
	descs := discoverBasicLits(t, src)

	var nonEmptyToEmpty int
	for _, d := range descs {
		if d.Kind == mutation.StringLitNonEmptyToEmpty {
			nonEmptyToEmpty++
			t.Logf("non_empty->empty: %s -> %s", d.Original, d.Replacement)
		}
	}

	if nonEmptyToEmpty != 2 {
		t.Errorf("expected 2 non_empty->empty mutations, got %d", nonEmptyToEmpty)
	}
}

func TestStringLiteralEmptyToSentinel(t *testing.T) {
	src := `package sample

var empty = ""
`
	descs := discoverBasicLits(t, src)

	var emptyToSentinel int
	for _, d := range descs {
		if d.Kind == mutation.StringLitEmptyToSentinel {
			emptyToSentinel++
			t.Logf("empty->sentinel: %s -> %s", d.Original, d.Replacement)
		}
	}

	if emptyToSentinel != 1 {
		t.Errorf("expected 1 empty->sentinel mutation, got %d", emptyToSentinel)
	}
}

func TestStringLiteralSkipsRawStrings(t *testing.T) {
	src := "package sample\n\nvar re = `[a-z]+`\n"
	descs := discoverBasicLits(t, src)

	for _, d := range descs {
		if d.Kind == mutation.StringLitNonEmptyToEmpty {
			t.Errorf("should not mutate raw string literals, got %s", d.Original)
		}
	}
}

func TestStringLiteralSkipsIntegers(t *testing.T) {
	src := `package sample

var n = 42
`
	descs := discoverBasicLits(t, src)

	if len(descs) != 0 {
		t.Errorf("expected no mutations for integer literals, got %d", len(descs))
	}
}
