package returnval_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/returnval"
)

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
			if nt != "*ast.ReturnStmt" {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if _, ok := n.(*ast.ReturnStmt); !ok {
					return true
				}
				all = slices.AppendSeq(all, s.Discover(ctx, n))
				return true
			})
		}
	}
	return all
}

func TestNilErrorDiscovery(t *testing.T) {
	src := `package sample

import "fmt"

func Validate(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return len(s), nil
}
`
	descs := discoverAll(t, src)

	var nilError []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.ReturnNilError {
			nilError = append(nilError, d)
		}
	}

	// Only the first return has a non-nil error.
	if len(nilError) != 1 {
		for _, d := range descs {
			t.Logf("  %s: %s -> %s", d.Kind, d.Original, d.Replacement)
		}
		t.Fatalf("expected 1 nil_error mutation, got %d", len(nilError))
	}

	d := nilError[0]
	if d.Replacement != "nil" {
		t.Errorf("expected replacement 'nil', got %q", d.Replacement)
	}
	t.Logf("nil_error: %s -> %s", d.Original, d.Replacement)
}

func TestNilErrorSkipsAlreadyNil(t *testing.T) {
	src := `package sample

func AlreadyNil() error {
	return nil
}
`
	descs := discoverAll(t, src)

	for _, d := range descs {
		if d.Kind == mutation.ReturnNilError {
			t.Error("should not generate nil_error for 'return nil'")
		}
	}
}

func TestZeroValueDiscovery(t *testing.T) {
	src := `package sample

import "fmt"

func Validate(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return len(s), nil
}
`
	descs := discoverAll(t, src)

	var zeroVal []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.ReturnZeroValue {
			zeroVal = append(zeroVal, d)
		}
	}

	// "return len(s), nil" — len(s) can be zeroed to 0.
	// "return 0, fmt.Errorf()" — 0 is already zero, skip.
	if len(zeroVal) != 1 {
		for _, d := range descs {
			t.Logf("  %s: %s -> %s", d.Kind, d.Original, d.Replacement)
		}
		t.Fatalf("expected 1 zero_value mutation, got %d", len(zeroVal))
	}

	d := zeroVal[0]
	if d.Replacement != "0" {
		t.Errorf("expected replacement '0', got %q", d.Replacement)
	}
	t.Logf("zero_value: %s -> %s", d.Original, d.Replacement)
}

func TestZeroValuePointerAndSlice(t *testing.T) {
	src := `package sample

func GetSlice() []int {
	return []int{1, 2, 3}
}

func GetPtr() *int {
	x := 42
	return &x
}
`
	descs := discoverAll(t, src)

	var zeroVal []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.ReturnZeroValue {
			zeroVal = append(zeroVal, d)
		}
	}

	if len(zeroVal) != 2 {
		for _, d := range descs {
			t.Logf("  %s: %s -> %s", d.Kind, d.Original, d.Replacement)
		}
		t.Fatalf("expected 2 zero_value mutations (slice+ptr), got %d", len(zeroVal))
	}

	for _, d := range zeroVal {
		if d.Replacement != "nil" {
			t.Errorf("expected 'nil' for pointer/slice zero, got %q", d.Replacement)
		}
		t.Logf("zero_value: %s -> %s", d.Original, d.Replacement)
	}
}

func TestNegateBoolDiscovery(t *testing.T) {
	src := `package sample

func IsValid(s string) bool {
	return len(s) > 0
}

func IsEmpty(s string) bool {
	return len(s) == 0
}
`
	descs := discoverAll(t, src)

	var negateBool []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.ReturnNegateBool {
			negateBool = append(negateBool, d)
		}
	}

	if len(negateBool) != 2 {
		for _, d := range descs {
			t.Logf("  %s: %s -> %s", d.Kind, d.Original, d.Replacement)
		}
		t.Fatalf("expected 2 negate_bool mutations, got %d", len(negateBool))
	}

	for _, d := range negateBool {
		t.Logf("negate_bool: %s -> %s", d.Original, d.Replacement)
	}
}

func TestNegateBoolSkipsAlreadyNegated(t *testing.T) {
	src := `package sample

func IsInvalid(s string) bool {
	return !(len(s) > 0)
}
`
	descs := discoverAll(t, src)

	for _, d := range descs {
		if d.Kind == mutation.ReturnNegateBool {
			t.Errorf("should not negate already-negated expression: %s", d.Original)
		}
	}
}

func TestNegateBoolMultiReturn(t *testing.T) {
	src := `package sample

func Lookup(key string) (string, bool) {
	if key == "found" {
		return "value", true
	}
	return "", false
}
`
	descs := discoverAll(t, src)

	var negateBool []mutation.Descriptor
	for _, d := range descs {
		if d.Kind == mutation.ReturnNegateBool {
			negateBool = append(negateBool, d)
		}
	}

	// "return "value", true" → negate true
	// "return "", false" → negate false
	if len(negateBool) != 2 {
		for _, d := range descs {
			t.Logf("  %s: %s -> %s", d.Kind, d.Original, d.Replacement)
		}
		t.Fatalf("expected 2 negate_bool mutations, got %d", len(negateBool))
	}
}
