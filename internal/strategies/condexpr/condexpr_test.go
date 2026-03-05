package condexpr_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/condexpr"
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

	var all []mutation.Descriptor
	for _, s := range strategy.All() {
		for _, nt := range s.NodeTypes() {
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				typeName := ""
				switch n.(type) {
				case *ast.IfStmt:
					typeName = "*ast.IfStmt"
				case *ast.ForStmt:
					typeName = "*ast.ForStmt"
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

func TestCondExprIfStatement(t *testing.T) {
	src := `package sample

func Check(x int) string {
	if x > 0 {
		return "positive"
	}
	return "non-positive"
}
`
	descs := discoverNodes(t, src)

	var condTrue, condFalse int
	for _, d := range descs {
		switch d.Kind {
		case mutation.ConditionalExprTrue:
			condTrue++
			t.Logf("cond->true: %s at line %d", d.Original, d.StartPos.Line)
		case mutation.ConditionalExprFalse:
			condFalse++
			t.Logf("cond->false: %s at line %d", d.Original, d.StartPos.Line)
		}
	}

	if condTrue != 1 {
		t.Errorf("expected 1 condition->true, got %d", condTrue)
	}
	if condFalse != 1 {
		t.Errorf("expected 1 condition->false, got %d", condFalse)
	}
}

func TestCondExprForLoop(t *testing.T) {
	src := `package sample

func Counter(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += i
	}
	return total
}
`
	descs := discoverNodes(t, src)

	var condTrue, condFalse int
	for _, d := range descs {
		switch d.Kind {
		case mutation.ConditionalExprTrue:
			condTrue++
		case mutation.ConditionalExprFalse:
			condFalse++
			t.Logf("for cond->false: %s", d.Original)
		}
	}

	// For loops only get false (not true, to avoid infinite loops).
	if condTrue != 0 {
		t.Errorf("expected 0 for-condition->true, got %d", condTrue)
	}
	if condFalse != 1 {
		t.Errorf("expected 1 for-condition->false, got %d", condFalse)
	}
}

func TestCondExprSkipsLiteralBool(t *testing.T) {
	src := `package sample

func Always() {
	if true {
		println("yes")
	}
}
`
	descs := discoverNodes(t, src)

	for _, d := range descs {
		if d.Kind == mutation.ConditionalExprTrue || d.Kind == mutation.ConditionalExprFalse {
			t.Errorf("should not mutate literal bool condition, got %s", d.Kind)
		}
	}
}
