package negatives

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&negativesStrategy{})
}

type negativesStrategy struct{}

func (s *negativesStrategy) Name() string        { return "negatives" }
func (s *negativesStrategy) NodeTypes() []string  { return []string{"*ast.UnaryExpr"} }

func (s *negativesStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.UnaryExpr)
		if !ok {
			return
		}

		if n.Op != token.SUB {
			return
		}

		pos := ctx.Fset.Position(n.OpPos)
		// Unary '-' is a single byte.
		endOffset := pos.Offset + 1

		desc := mutation.Descriptor{
			File:     ctx.FilePath,
			PkgPath:  ctx.PkgPath,
			StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
			EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + 1, Offset: endOffset},
			Kind:     mutation.NegativesRemove,
			Status:   mutation.Runnable,
			Original:    "-",
			Replacement: "+",
			Apply: mutation.ApplySpec{
				TokenSwap: &mutation.TokenSwapSpec{
					OriginalToken:    "-",
					ReplacementToken: "+",
					StartOffset:      pos.Offset,
					EndOffset:        endOffset,
				},
			},
		}

		yield(desc)
	}
}
