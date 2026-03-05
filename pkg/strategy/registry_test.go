package strategy

import (
	"go/ast"
	"iter"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

type testStrategy struct {
	name      string
	nodeTypes []string
}

func (s *testStrategy) Name() string        { return s.name }
func (s *testStrategy) NodeTypes() []string  { return s.nodeTypes }
func (s *testStrategy) Discover(_ *DiscoveryContext, _ ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {}
}

func TestRegisterAndLookup(t *testing.T) {
	// Save and restore state.
	origRegistry := registry
	origByNode := byNode
	t.Cleanup(func() {
		mu.Lock()
		registry = origRegistry
		byNode = origByNode
		mu.Unlock()
	})

	mu.Lock()
	registry = nil
	byNode = make(map[string][]Strategy)
	mu.Unlock()

	s := &testStrategy{name: "test/one", nodeTypes: []string{"*ast.BinaryExpr"}}
	Register(s)

	all := All()
	if len(all) != 1 {
		t.Fatalf("All() = %d strategies, want 1", len(all))
	}
	if all[0].Name() != "test/one" {
		t.Errorf("All()[0].Name() = %q, want %q", all[0].Name(), "test/one")
	}

	byType := ForNodeType("*ast.BinaryExpr")
	if len(byType) != 1 {
		t.Fatalf("ForNodeType(*ast.BinaryExpr) = %d, want 1", len(byType))
	}

	byType = ForNodeType("*ast.IfStmt")
	if len(byType) != 0 {
		t.Fatalf("ForNodeType(*ast.IfStmt) = %d, want 0", len(byType))
	}
}

func TestRegisterDuplicate(t *testing.T) {
	origRegistry := registry
	origByNode := byNode
	t.Cleanup(func() {
		mu.Lock()
		registry = origRegistry
		byNode = origByNode
		mu.Unlock()
	})

	mu.Lock()
	registry = nil
	byNode = make(map[string][]Strategy)
	mu.Unlock()

	s := &testStrategy{name: "test/dup", nodeTypes: []string{"*ast.BinaryExpr"}}
	Register(s)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	Register(s)
}
