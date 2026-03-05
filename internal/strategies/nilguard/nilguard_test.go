package nilguard_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/nilguard"
)

func discoverIfStmts(t *testing.T, src string) []mutation.Descriptor {
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
			if nt != "*ast.IfStmt" {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if _, ok := n.(*ast.IfStmt); !ok {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func nilGuardDescs(descs []mutation.Descriptor) []mutation.Descriptor {
	var result []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.NilGuardRemove {
			result = append(result, d)
		}
	}
	return result
}

func TestNilGuardNotNil(t *testing.T) {
	src := `package sample

type T struct{ V int }

func SafeGet(p *T) int {
	if p != nil {
		return p.V
	}
	return 0
}
`
	descs := nilGuardDescs(discoverIfStmts(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 nilguard mutation, got %d", len(descs))
	}

	d := descs[0]
	if d.Replacement != "true" {
		t.Errorf("expected replacement 'true', got %q", d.Replacement)
	}
	t.Logf("nilguard: %s -> %s", d.Original, d.Replacement)
}

func TestNilGuardEqNil(t *testing.T) {
	src := `package sample

func Guard(p *int) int {
	if p == nil {
		return -1
	}
	return *p
}
`
	descs := nilGuardDescs(discoverIfStmts(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 nilguard mutation, got %d", len(descs))
	}

	d := descs[0]
	if d.Replacement != "false" {
		t.Errorf("expected replacement 'false', got %q", d.Replacement)
	}
	t.Logf("nilguard: %s -> %s", d.Original, d.Replacement)
}

func TestNilGuardSkipsNonNil(t *testing.T) {
	src := `package sample

func Check(x int) {
	if x > 0 {
		println("positive")
	}
}
`
	descs := nilGuardDescs(discoverIfStmts(t, src))

	if len(descs) != 0 {
		t.Errorf("expected 0 nilguard mutations for non-nil check, got %d", len(descs))
	}
}

func TestNilGuardErrorPattern(t *testing.T) {
	src := `package sample

import "fmt"

func Parse(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return len(s), nil
}
`
	descs := nilGuardDescs(discoverIfStmts(t, src))

	// s == "" is not a nil check
	if len(descs) != 0 {
		t.Errorf("expected 0 nilguard mutations for string comparison, got %d", len(descs))
	}
}

func TestNilGuardInterfaceNil(t *testing.T) {
	src := `package sample

type Handler interface{ Handle() }

func Run(h Handler) {
	if h != nil {
		h.Handle()
	}
}
`
	descs := nilGuardDescs(discoverIfStmts(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 nilguard mutation, got %d", len(descs))
	}
	if descs[0].Replacement != "true" {
		t.Errorf("expected 'true', got %q", descs[0].Replacement)
	}
}
