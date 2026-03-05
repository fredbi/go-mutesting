package walker

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"reflect"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

// Walker performs a single AST walk dispatching to registered strategies.
type Walker struct {
	strategies []strategy.Strategy
}

// New creates a Walker using the given strategies.
// If strategies is nil, all registered strategies are used.
func New(strategies []strategy.Strategy) *Walker {
	if strategies == nil {
		strategies = strategy.All()
	}
	return &Walker{strategies: strategies}
}

// Discover performs a single ast.Inspect pass over the file and lazily yields
// mutation descriptors from all matching strategies.
func (w *Walker) Discover(
	fset *token.FileSet,
	file *ast.File,
	pkg *types.Package,
	info *types.Info,
	src []byte,
	filePath, pkgPath string,
) iter.Seq[mutation.Descriptor] {
	ctx := &strategy.DiscoveryContext{
		Fset:     fset,
		File:     file,
		Pkg:      pkg,
		Info:     info,
		Src:      src,
		FilePath: filePath,
		PkgPath:  pkgPath,
	}

	fnResolver := buildFuncNameResolver(fset, file)

	// Build a dispatch map: nodeType -> []Strategy
	dispatch := make(map[string][]strategy.Strategy)
	for _, s := range w.strategies {
		for _, nt := range s.NodeTypes() {
			dispatch[nt] = append(dispatch[nt], s)
		}
	}

	return func(yield func(mutation.Descriptor) bool) {
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return false
			}

			nodeType := reflect.TypeOf(n).String()
			strategies := dispatch[nodeType]
			if len(strategies) == 0 {
				return true
			}

			for _, s := range strategies {
				for desc := range s.Discover(ctx, n) {
					// Fill in function name from the resolver.
					if desc.FuncName == "" {
						desc.FuncName = fnResolver.resolve(desc.StartPos.Offset)
					}

					// Compute stable ID if not already set.
					if desc.ID == "" {
						desc.ID = mutation.ComputeID(
							desc.File, desc.Kind,
							desc.StartPos.Line, desc.StartPos.Column,
							desc.Original, desc.Replacement,
						)
					}

					if !yield(desc) {
						return false
					}
				}
			}

			return true
		})
	}
}
