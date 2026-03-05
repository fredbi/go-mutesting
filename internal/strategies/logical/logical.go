package logical

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&logicalStrategy{})
}

var mutations = []struct {
	from token.Token
	to   token.Token
	kind mutation.Kind
}{
	{token.LAND, token.LOR, mutation.LogicalAndToOr},
	{token.LOR, token.LAND, mutation.LogicalOrToAnd},
}

type logicalStrategy struct{}

func (s *logicalStrategy) Name() string        { return "logical" }
func (s *logicalStrategy) NodeTypes() []string  { return []string{"*ast.BinaryExpr"} }

func (s *logicalStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.BinaryExpr)
		if !ok {
			return
		}

		for _, m := range mutations {
			if n.Op != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.OpPos)
			endOffset := pos.Offset + len(n.Op.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Op.String()), Offset: endOffset},
				Kind:     m.kind,
				Status:   mutation.Runnable,
				Original:    m.from.String(),
				Replacement: m.to.String(),
				Apply: mutation.ApplySpec{
					TokenSwap: &mutation.TokenSwapSpec{
						OriginalToken:    m.from.String(),
						ReplacementToken: m.to.String(),
						StartOffset:      pos.Offset,
						EndOffset:        endOffset,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}
