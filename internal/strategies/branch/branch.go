package branch

import (
	"fmt"
	"go/ast"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&emptyIfStrategy{})
	strategy.Register(&emptyElseStrategy{})
	strategy.Register(&emptyCaseStrategy{})
}

// emptyIfStrategy replaces if-body with a noop.
type emptyIfStrategy struct{}

func (s *emptyIfStrategy) Name() string        { return "branch/empty_if" }
func (s *emptyIfStrategy) NodeTypes() []string  { return []string{"*ast.IfStmt"} }

func (s *emptyIfStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.IfStmt)
		if !ok {
			return
		}

		bodyStart := ctx.Fset.Position(n.Body.Lbrace)
		bodyEnd := ctx.Fset.Position(n.Body.Rbrace)
		stmtCount := len(n.Body.List)

		desc := mutation.Descriptor{
			File:     ctx.FilePath,
			PkgPath:  ctx.PkgPath,
			StartPos: mutation.Position{Line: bodyStart.Line, Column: bodyStart.Column, Offset: bodyStart.Offset},
			EndPos:   mutation.Position{Line: bodyEnd.Line, Column: bodyEnd.Column, Offset: bodyEnd.Offset},
			Kind:     mutation.BranchEmptyIf,
			Status:   mutation.Runnable,
			Original:    fmt.Sprintf("if body: %d statements", stmtCount),
			Replacement: "empty block",
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "IfStmt",
					Action:      mutation.ActionEmptyBlock,
					TargetIndex: -1,
					StartOffset: bodyStart.Offset,
					EndOffset:   bodyEnd.Offset + 1, // include the closing brace
				},
			},
		}

		yield(desc)
	}
}

// emptyElseStrategy replaces else-body with a noop.
type emptyElseStrategy struct{}

func (s *emptyElseStrategy) Name() string        { return "branch/empty_else" }
func (s *emptyElseStrategy) NodeTypes() []string  { return []string{"*ast.IfStmt"} }

func (s *emptyElseStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.IfStmt)
		if !ok {
			return
		}

		// No else clause.
		if n.Else == nil {
			return
		}

		// Skip else-if chains (only target plain else blocks).
		block, ok := n.Else.(*ast.BlockStmt)
		if !ok {
			return
		}

		elseStart := ctx.Fset.Position(block.Lbrace)
		elseEnd := ctx.Fset.Position(block.Rbrace)
		stmtCount := len(block.List)

		desc := mutation.Descriptor{
			File:     ctx.FilePath,
			PkgPath:  ctx.PkgPath,
			StartPos: mutation.Position{Line: elseStart.Line, Column: elseStart.Column, Offset: elseStart.Offset},
			EndPos:   mutation.Position{Line: elseEnd.Line, Column: elseEnd.Column, Offset: elseEnd.Offset},
			Kind:     mutation.BranchEmptyElse,
			Status:   mutation.Runnable,
			Original:    fmt.Sprintf("else body: %d statements", stmtCount),
			Replacement: "empty block",
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "IfStmt.Else",
					Action:      mutation.ActionEmptyBlock,
					TargetIndex: -1,
					StartOffset: elseStart.Offset,
					EndOffset:   elseEnd.Offset + 1,
				},
			},
		}

		yield(desc)
	}
}

// emptyCaseStrategy replaces case body with a noop.
type emptyCaseStrategy struct{}

func (s *emptyCaseStrategy) Name() string        { return "branch/empty_case" }
func (s *emptyCaseStrategy) NodeTypes() []string  { return []string{"*ast.CaseClause"} }

func (s *emptyCaseStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.CaseClause)
		if !ok {
			return
		}

		if len(n.Body) == 0 {
			return
		}

		colonPos := ctx.Fset.Position(n.Colon)
		lastStmt := n.Body[len(n.Body)-1]
		bodyEnd := ctx.Fset.Position(lastStmt.End())

		desc := mutation.Descriptor{
			File:     ctx.FilePath,
			PkgPath:  ctx.PkgPath,
			StartPos: mutation.Position{Line: colonPos.Line, Column: colonPos.Column, Offset: colonPos.Offset},
			EndPos:   mutation.Position{Line: bodyEnd.Line, Column: bodyEnd.Column, Offset: bodyEnd.Offset},
			Kind:     mutation.BranchEmptyCase,
			Status:   mutation.Runnable,
			Original:    fmt.Sprintf("case body: %d statements", len(n.Body)),
			Replacement: "empty block",
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "CaseClause",
					Action:      mutation.ActionEmptyBlock,
					TargetIndex: -1,
					StartOffset: colonPos.Offset,
					EndOffset:   bodyEnd.Offset,
				},
			},
		}

		yield(desc)
	}
}
