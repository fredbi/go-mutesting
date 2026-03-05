package panicremove_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/panicremove"
)

func discoverExprStmts(t *testing.T, src string) []mutation.Descriptor {
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
			if nt != "*ast.ExprStmt" {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if _, ok := n.(*ast.ExprStmt); !ok {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func TestPanicToReturnWithValues(t *testing.T) {
	src := `package sample

func MustParse(s string) int {
	if s == "" {
		panic("empty input")
	}
	return len(s)
}
`
	descs := discoverExprStmts(t, src)

	var panicMuts []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.PanicToReturn {
			panicMuts = append(panicMuts, d)
		}
	}

	if len(panicMuts) != 1 {
		t.Fatalf("expected 1 panic->return mutation, got %d", len(panicMuts))
	}

	d := panicMuts[0]
	if d.Replacement != "return 0" {
		t.Errorf("expected 'return 0', got %q", d.Replacement)
	}
	t.Logf("panic->return: %s -> %s", d.Original, d.Replacement)
}

func TestPanicToReturnMultipleValues(t *testing.T) {
	src := `package sample

func MustLookup(key string) (string, error) {
	panic("not implemented")
}
`
	descs := discoverExprStmts(t, src)

	var panicMuts []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.PanicToReturn {
			panicMuts = append(panicMuts, d)
		}
	}

	if len(panicMuts) != 1 {
		t.Fatalf("expected 1 panic->return mutation, got %d", len(panicMuts))
	}

	d := panicMuts[0]
	if d.Replacement != `return "", nil` {
		t.Errorf("expected 'return \"\", nil', got %q", d.Replacement)
	}
	t.Logf("panic->return: %s -> %s", d.Original, d.Replacement)
}

func TestPanicToReturnVoid(t *testing.T) {
	src := `package sample

func PanicVoid() {
	panic("fatal")
}
`
	descs := discoverExprStmts(t, src)

	var panicMuts []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.PanicToReturn {
			panicMuts = append(panicMuts, d)
		}
	}

	if len(panicMuts) != 1 {
		t.Fatalf("expected 1 panic->return mutation for void func, got %d", len(panicMuts))
	}

	d := panicMuts[0]
	if d.Replacement != "return" {
		t.Errorf("expected 'return', got %q", d.Replacement)
	}
}

func TestPanicSkipsNonBuiltin(t *testing.T) {
	// A user-defined function named "panic" should not trigger this.
	src := `package sample

func myPanic(msg string) {
	println(msg)
}

func Test() {
	var panic = myPanic
	panic("oops")
}
`
	descs := discoverExprStmts(t, src)

	for _, d := range descs {
		if d.Kind == mutation.PanicToReturn {
			t.Error("should not mutate user-defined 'panic' function")
		}
	}
}
