package expression

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&removeTermStrategy{})
}

// removeTermStrategy replaces operands of && with true and || with false.
// Ported from go-mutesting's expression/remove mutator.
type removeTermStrategy struct{}

func (s *removeTermStrategy) Name() string        { return "expression/remove_term" }
func (s *removeTermStrategy) NodeTypes() []string  { return []string{"*ast.BinaryExpr"} }

func (s *removeTermStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.BinaryExpr)
		if !ok {
			return
		}

		if n.Op != token.LAND && n.Op != token.LOR {
			return
		}

		var replacement string
		var action mutation.StructuralAction
		switch n.Op {
		case token.LAND:
			replacement = "true"
			action = mutation.ActionReplaceWithTrue
		case token.LOR:
			replacement = "false"
			action = mutation.ActionReplaceWithFalse
		}

		// Mutation 1: replace left operand.
		xStart := ctx.Fset.Position(n.X.Pos())
		xEnd := ctx.Fset.Position(n.X.End())

		descX := mutation.Descriptor{
			File:     ctx.FilePath,
			PkgPath:  ctx.PkgPath,
			StartPos: mutation.Position{Line: xStart.Line, Column: xStart.Column, Offset: xStart.Offset},
			EndPos:   mutation.Position{Line: xEnd.Line, Column: xEnd.Column, Offset: xEnd.Offset},
			Kind:     mutation.ExpressionRemoveTerm,
			Status:   mutation.Runnable,
			Original:    "left operand of " + n.Op.String(),
			Replacement: replacement,
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "BinaryExpr.X",
					Action:      action,
					TargetIndex: -1,
					StartOffset: xStart.Offset,
					EndOffset:   xEnd.Offset,
				},
			},
		}

		if !yield(descX) {
			return
		}

		// Mutation 2: replace right operand.
		yStart := ctx.Fset.Position(n.Y.Pos())
		yEnd := ctx.Fset.Position(n.Y.End())

		descY := mutation.Descriptor{
			File:     ctx.FilePath,
			PkgPath:  ctx.PkgPath,
			StartPos: mutation.Position{Line: yStart.Line, Column: yStart.Column, Offset: yStart.Offset},
			EndPos:   mutation.Position{Line: yEnd.Line, Column: yEnd.Column, Offset: yEnd.Offset},
			Kind:     mutation.ExpressionRemoveTerm,
			Status:   mutation.Runnable,
			Original:    "right operand of " + n.Op.String(),
			Replacement: replacement,
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "BinaryExpr.Y",
					Action:      action,
					TargetIndex: -1,
					StartOffset: yStart.Offset,
					EndOffset:   yEnd.Offset,
				},
			},
		}

		yield(descY)
	}
}
