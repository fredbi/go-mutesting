package loopctrl

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&loopCtrlStrategy{})
}

var mutations = []struct {
	from token.Token
	to   token.Token
	kind mutation.Kind
}{
	{token.BREAK, token.CONTINUE, mutation.LoopCtrlBreakToContinue},
	{token.CONTINUE, token.BREAK, mutation.LoopCtrlContinueToBreak},
}

type loopCtrlStrategy struct{}

func (s *loopCtrlStrategy) Name() string        { return "loopctrl" }
func (s *loopCtrlStrategy) NodeTypes() []string  { return []string{"*ast.BranchStmt"} }

func (s *loopCtrlStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.BranchStmt)
		if !ok {
			return
		}

		// Only mutate simple break/continue (no label).
		if n.Label != nil {
			return
		}

		for _, m := range mutations {
			if n.Tok != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.TokPos)
			endOffset := pos.Offset + len(n.Tok.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Tok.String()), Offset: endOffset},
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
