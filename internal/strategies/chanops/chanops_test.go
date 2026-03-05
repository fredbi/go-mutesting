package chanops_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/chanops"
)

func discoverBlocks(t *testing.T, src string) []mutation.Descriptor {
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
		hasTarget := false
		for _, nt := range s.NodeTypes() {
			if nt == "*ast.BlockStmt" || nt == "*ast.CaseClause" {
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
			case *ast.BlockStmt, *ast.CaseClause:
				all = slices.AppendSeq(all, s.Discover(ctx, n))
			}
			return true
		})
	}
	return all
}

func chanDescs(descs []mutation.Descriptor) []mutation.Descriptor {
	var result []mutation.Descriptor
	for _, d := range descs {
		if d.Kind.Category() == "chanops" {
			result = append(result, d)
		}
	}
	return result
}

func TestRemoveClose(t *testing.T) {
	src := `package sample

func Producer(ch chan int) {
	ch <- 1
	ch <- 2
	close(ch)
}
`
	descs := chanDescs(discoverBlocks(t, src))

	var hasClose bool
	var hasSend int
	for _, d := range descs {
		switch d.Kind {
		case mutation.ChanOpsRemoveClose:
			hasClose = true
			t.Logf("remove close: %s (MayHang=%v)", d.Original, d.Kind.MayHang())
		case mutation.ChanOpsRemoveSend:
			hasSend++
			t.Logf("remove send: %s (MayHang=%v)", d.Original, d.Kind.MayHang())
		}
	}

	if !hasClose {
		t.Error("missing close removal mutation")
	}
	if hasSend != 2 {
		t.Errorf("expected 2 send removals, got %d", hasSend)
	}
}

func TestRemoveDeferClose(t *testing.T) {
	src := `package sample

func Producer2(ch chan int) {
	defer close(ch)
	ch <- 1
}
`
	descs := chanDescs(discoverBlocks(t, src))

	var hasClose bool
	for _, d := range descs {
		if d.Kind == mutation.ChanOpsRemoveClose {
			hasClose = true
			t.Logf("remove defer close: %s", d.Original)
		}
	}

	if !hasClose {
		t.Error("missing defer close removal mutation")
	}
}

func TestRemoveReceive(t *testing.T) {
	src := `package sample

func Drain(ch chan int) {
	<-ch
}
`
	descs := chanDescs(discoverBlocks(t, src))

	var hasReceive bool
	for _, d := range descs {
		if d.Kind == mutation.ChanOpsRemoveReceive {
			hasReceive = true
			t.Logf("remove receive: %s (MayHang=%v)", d.Original, d.Kind.MayHang())
		}
	}

	if !hasReceive {
		t.Error("missing receive removal mutation")
	}
}

func TestMayHangClassification(t *testing.T) {
	if !mutation.ChanOpsRemoveClose.MayHang() {
		t.Error("remove close SHOULD be MayHang")
	}
	if !mutation.ChanOpsRemoveSend.MayHang() {
		t.Error("remove send SHOULD be MayHang")
	}
	if !mutation.ChanOpsRemoveReceive.MayHang() {
		t.Error("remove receive SHOULD be MayHang")
	}
}

func TestSkipsUserDefinedClose(t *testing.T) {
	src := `package sample

func myClose(x int) {}

func Test() {
	var close = myClose
	close(42)
}
`
	descs := chanDescs(discoverBlocks(t, src))

	for _, d := range descs {
		if d.Kind == mutation.ChanOpsRemoveClose {
			t.Error("should not match user-defined 'close' function")
		}
	}
}
