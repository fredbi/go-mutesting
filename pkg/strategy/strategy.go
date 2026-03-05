package strategy

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// Strategy defines the interface for mutation discovery strategies.
//
// Strategies must be stateless and goroutine-safe (pure functions of context + node).
type Strategy interface {
	// Name returns the unique name of this strategy.
	Name() string

	// NodeTypes returns the AST node types this strategy handles,
	// e.g. ["*ast.BinaryExpr"]. Used for dispatch optimization.
	NodeTypes() []string

	// Discover yields mutation descriptors for the given AST node.
	Discover(ctx *DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor]
}

// DiscoveryContext provides the full context needed by strategies for mutation discovery.
type DiscoveryContext struct {
	Fset     *token.FileSet
	File     *ast.File
	Pkg      *types.Package
	Info     *types.Info
	Src      []byte // Original file bytes (for offset computation)
	FilePath string // Absolute path to the file
	PkgPath  string // Go import path
}
