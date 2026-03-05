package argswap_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/argswap"
)

func discoverCallExprs(t *testing.T, src string) []mutation.Descriptor {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	conf := types.Config{Error: func(err error) {}}
	pkg, _ := conf.Check("sample", fset, []*ast.File{file}, info)

	ctx := &strategy.DiscoveryContext{
		Fset: fset, File: file, Pkg: pkg, Info: info,
		Src: []byte(src), FilePath: "test.go", PkgPath: "sample",
	}

	var all []mutation.Descriptor
	for _, s := range strategy.All() {
		for _, nt := range s.NodeTypes() {
			if nt != "*ast.CallExpr" {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if _, ok := n.(*ast.CallExpr); !ok {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func argSwapDescs(descs []mutation.Descriptor) []mutation.Descriptor {
	var result []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.ArgSwap {
			result = append(result, d)
		}
	}
	return result
}

func TestSwapAdjacentSameType(t *testing.T) {
	src := `package sample

func add(a, b int) int { return a + b }

func use() int {
	return add(1, 2)
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 argswap mutation, got %d", len(descs))
	}

	d := descs[0]
	if d.Original != "1, 2" {
		t.Errorf("expected original '1, 2', got %q", d.Original)
	}
	if d.Replacement != "2, 1" {
		t.Errorf("expected replacement '2, 1', got %q", d.Replacement)
	}
	t.Logf("argswap: %s -> %s", d.Original, d.Replacement)
}

func TestSwapDifferentTypes(t *testing.T) {
	src := `package sample

func mixed(a int, b string) {}

func use() {
	mixed(1, "hello")
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 0 {
		t.Errorf("expected 0 argswap mutations for different types, got %d", len(descs))
	}
}

func TestSwapThreeSameType(t *testing.T) {
	src := `package sample

func triple(a, b, c int) int { return a + b + c }

func use() int {
	return triple(1, 2, 3)
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	// Adjacent pairs: (1,2) and (2,3) => 2 mutations.
	if len(descs) != 2 {
		t.Fatalf("expected 2 argswap mutations, got %d", len(descs))
	}

	if descs[0].Original != "1, 2" || descs[0].Replacement != "2, 1" {
		t.Errorf("first swap: expected '1, 2' -> '2, 1', got %q -> %q", descs[0].Original, descs[0].Replacement)
	}
	if descs[1].Original != "2, 3" || descs[1].Replacement != "3, 2" {
		t.Errorf("second swap: expected '2, 3' -> '3, 2', got %q -> %q", descs[1].Original, descs[1].Replacement)
	}
}

func TestSwapMixedTypes(t *testing.T) {
	// Only a and b are int (adjacent), flag is bool => 1 swap.
	src := `package sample

func f(a, b int, flag bool) {}

func use() {
	f(1, 2, true)
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 argswap mutation, got %d", len(descs))
	}
	t.Logf("argswap: %s -> %s", descs[0].Original, descs[0].Replacement)
}

func TestSwapSkipsIdenticalArgs(t *testing.T) {
	src := `package sample

func add(a, b int) int { return a + b }

func use() int {
	x := 5
	return add(x, x)
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 0 {
		t.Errorf("expected 0 mutations for identical args, got %d", len(descs))
	}
}

func TestSwapSkipsVariadicExpansion(t *testing.T) {
	src := `package sample

func variadic(args ...int) int { return 0 }

func use() int {
	s := []int{1, 2}
	return variadic(s...)
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 0 {
		t.Errorf("expected 0 mutations for variadic expansion, got %d", len(descs))
	}
}

func TestSwapMethodCall(t *testing.T) {
	src := `package sample

type Calc struct{}

func (c Calc) Add(a, b int) int { return a + b }

func use() int {
	c := Calc{}
	return c.Add(3, 7)
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 argswap mutation for method call, got %d", len(descs))
	}
	if descs[0].Original != "3, 7" || descs[0].Replacement != "7, 3" {
		t.Errorf("expected '3, 7' -> '7, 3', got %q -> %q", descs[0].Original, descs[0].Replacement)
	}
}

func TestSwapStringArgs(t *testing.T) {
	src := `package sample

func concat(a, b string) string { return a + b }

func use() string {
	return concat("hello", "world")
}
`
	descs := argSwapDescs(discoverCallExprs(t, src))

	if len(descs) != 1 {
		t.Fatalf("expected 1 argswap mutation, got %d", len(descs))
	}
	t.Logf("argswap: %s -> %s", descs[0].Original, descs[0].Replacement)
}
