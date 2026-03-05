package walker_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/walker"

	// Register all strategies.
	_ "github.com/fredbi/go-mutesting/internal/strategies/all"
)

func parseAndTypeCheck(t *testing.T, path string) (*token.FileSet, *ast.File, *types.Package, *types.Info, []byte) {
	t.Helper()

	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	info := &types.Info{
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	conf := types.Config{Error: func(err error) {}}
	pkg, _ := conf.Check("sample", fset, []*ast.File{file}, info)

	return fset, file, pkg, info, src
}

func TestWalkerDiscover(t *testing.T) {
	const testFile = "../../testdata/sample.go"

	fset, file, pkg, info, src := parseAndTypeCheck(t, testFile)

	w := walker.New(nil) // use all registered strategies
	descriptors := slices.Collect(w.Discover(fset, file, pkg, info, src, testFile, "sample"))

	if len(descriptors) == 0 {
		t.Fatal("expected at least one mutation descriptor")
	}

	// Verify all descriptors have required fields.
	for i, d := range descriptors {
		if d.ID == "" {
			t.Errorf("descriptor[%d]: empty ID", i)
		}
		if d.File == "" {
			t.Errorf("descriptor[%d]: empty File", i)
		}
		if d.Kind == "" {
			t.Errorf("descriptor[%d]: empty Kind", i)
		}
		if d.Original == "" {
			t.Errorf("descriptor[%d]: empty Original", i)
		}
		if d.Replacement == "" {
			t.Errorf("descriptor[%d]: empty Replacement", i)
		}
		if d.Apply.TokenSwap == nil && d.Apply.Structural == nil {
			t.Errorf("descriptor[%d]: no ApplySpec", i)
		}
	}

	// Verify we get mutations from both token-swap and structural strategies.
	categories := make(map[string]bool)
	for _, d := range descriptors {
		categories[d.Kind.Category()] = true
	}

	expectedCategories := []string{"arithmetic", "conditional_boundary", "conditional_negation", "branch", "statement", "logical"}
	for _, cat := range expectedCategories {
		if !categories[cat] {
			t.Errorf("expected mutations in category %q, not found", cat)
		}
	}

	t.Logf("total mutations: %d, categories: %v", len(descriptors), categories)
}

func TestWalkerFuncNames(t *testing.T) {
	const testFile = "../../testdata/sample.go"

	fset, file, pkg, info, src := parseAndTypeCheck(t, testFile)

	w := walker.New(nil)
	descriptors := slices.Collect(w.Discover(fset, file, pkg, info, src, testFile, "sample"))

	funcNames := make(map[string]bool)
	for _, d := range descriptors {
		if d.FuncName != "" {
			funcNames[d.FuncName] = true
		}
	}

	expectedFuncs := []string{"Add", "Compare", "Logic", "Counter", "Negative", "SwitchCase", "Loop"}
	for _, fn := range expectedFuncs {
		if !funcNames[fn] {
			t.Errorf("expected mutations in function %q, not found (have: %v)", fn, funcNames)
		}
	}
}

func TestDescriptorUniqueness(t *testing.T) {
	const testFile = "../../testdata/sample.go"

	fset, file, pkg, info, src := parseAndTypeCheck(t, testFile)

	w := walker.New(nil)
	descriptors := slices.Collect(w.Discover(fset, file, pkg, info, src, testFile, "sample"))

	ids := make(map[string]int)
	for _, d := range descriptors {
		ids[d.ID]++
	}

	for id, count := range ids {
		if count > 1 {
			t.Errorf("duplicate ID %s (count=%d)", id, count)
		}
	}
}
