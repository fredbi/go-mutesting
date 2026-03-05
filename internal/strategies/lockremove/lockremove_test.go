package lockremove_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"

	_ "github.com/fredbi/go-mutesting/internal/strategies/lockremove"
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

func lockDescs(descs []mutation.Descriptor) []mutation.Descriptor {
	var result []mutation.Descriptor
	for _, d := range descs {
		if d.Kind.Category() == "lockremove" {
			result = append(result, d)
		}
	}
	return result
}

func TestRemoveLockUnlock(t *testing.T) {
	src := `package sample

import "sync"

func SafeIncrement(mu *sync.Mutex, counter *int) {
	mu.Lock()
	*counter++
	mu.Unlock()
}
`
	descs := lockDescs(discoverBlocks(t, src))

	if len(descs) != 2 {
		t.Fatalf("expected 2 lock mutations, got %d", len(descs))
	}

	var hasLock, hasUnlock bool
	for _, d := range descs {
		switch d.Kind {
		case mutation.LockRemoveLock:
			hasLock = true
			t.Logf("remove Lock: %s", d.Original)
		case mutation.LockRemoveUnlock:
			hasUnlock = true
			t.Logf("remove Unlock: %s (MayHang=%v)", d.Original, d.Kind.MayHang())
		}
	}

	if !hasLock {
		t.Error("missing Lock removal mutation")
	}
	if !hasUnlock {
		t.Error("missing Unlock removal mutation")
	}
}

func TestRemoveDeferUnlock(t *testing.T) {
	src := `package sample

import "sync"

func SafeRead(mu *sync.Mutex, data *string) string {
	mu.Lock()
	defer mu.Unlock()
	return *data
}
`
	descs := lockDescs(discoverBlocks(t, src))

	if len(descs) != 2 {
		t.Fatalf("expected 2 lock mutations (Lock + defer Unlock), got %d", len(descs))
	}

	var hasDeferUnlock bool
	for _, d := range descs {
		if d.Kind == mutation.LockRemoveUnlock {
			hasDeferUnlock = true
			t.Logf("remove defer Unlock: %s (MayHang=%v)", d.Original, d.Kind.MayHang())
		}
	}

	if !hasDeferUnlock {
		t.Error("missing defer Unlock removal mutation")
	}
}

func TestRemoveRLockRUnlock(t *testing.T) {
	src := `package sample

import "sync"

func SafeRead2(mu *sync.RWMutex, data *string) string {
	mu.RLock()
	defer mu.RUnlock()
	return *data
}
`
	descs := lockDescs(discoverBlocks(t, src))

	if len(descs) != 2 {
		t.Fatalf("expected 2 lock mutations (RLock + defer RUnlock), got %d", len(descs))
	}

	var hasRLock, hasRUnlock bool
	for _, d := range descs {
		switch d.Kind {
		case mutation.LockRemoveRLock:
			hasRLock = true
		case mutation.LockRemoveRUnlock:
			hasRUnlock = true
		}
	}

	if !hasRLock {
		t.Error("missing RLock removal")
	}
	if !hasRUnlock {
		t.Error("missing RUnlock removal")
	}
}

func TestMayHangClassification(t *testing.T) {
	if mutation.LockRemoveLock.MayHang() {
		t.Error("remove Lock should NOT be MayHang (it causes panic, not hang)")
	}
	if !mutation.LockRemoveUnlock.MayHang() {
		t.Error("remove Unlock SHOULD be MayHang (causes deadlock)")
	}
	if mutation.LockRemoveRLock.MayHang() {
		t.Error("remove RLock should NOT be MayHang")
	}
	if !mutation.LockRemoveRUnlock.MayHang() {
		t.Error("remove RUnlock SHOULD be MayHang")
	}
}

func TestSkipsNonLockMethods(t *testing.T) {
	src := `package sample

type MyService struct{}

func (s *MyService) Lock() {}
func (s *MyService) Process() {}

func Use(s *MyService) {
	s.Lock()
	s.Process()
}
`
	descs := lockDescs(discoverBlocks(t, src))

	// s.Lock() matches by name — intentional (Lock is a very distinctive name).
	// s.Process() should NOT match.
	if len(descs) != 1 {
		t.Fatalf("expected 1 lock mutation (Lock only), got %d", len(descs))
	}
	if descs[0].Kind != mutation.LockRemoveLock {
		t.Errorf("expected LockRemoveLock, got %s", descs[0].Kind)
	}
}
