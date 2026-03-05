package condexpr

import (
	"go/ast"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&condExprStrategy{})
}

// condExprStrategy replaces conditions in if/for statements with true/false.
//
// For if statements: generates both true and false replacements.
// For for-loop conditions: generates only false (replacing with true creates an infinite loop).
//
// Skips conditions that are already boolean literals.
type condExprStrategy struct{}

func (s *condExprStrategy) Name() string { return "conditional_expr" }
func (s *condExprStrategy) NodeTypes() []string {
	return []string{"*ast.IfStmt", "*ast.ForStmt"}
}

func (s *condExprStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		switch n := node.(type) {
		case *ast.IfStmt:
			discoverIfCond(ctx, n, yield)
		case *ast.ForStmt:
			discoverForCond(ctx, n, yield)
		}
	}
}

func discoverIfCond(ctx *strategy.DiscoveryContext, n *ast.IfStmt, yield func(mutation.Descriptor) bool) {
	if n.Cond == nil {
		return
	}

	if isBoolLiteral(n.Cond) {
		return
	}

	condStart := ctx.Fset.Position(n.Cond.Pos())
	condEnd := ctx.Fset.Position(n.Cond.End())
	original := sourceText(ctx.Src, condStart.Offset, condEnd.Offset)

	// Replace with true.
	descTrue := mutation.Descriptor{
		File:    ctx.FilePath,
		PkgPath: ctx.PkgPath,
		StartPos: mutation.Position{
			Line: condStart.Line, Column: condStart.Column, Offset: condStart.Offset,
		},
		EndPos: mutation.Position{
			Line: condEnd.Line, Column: condEnd.Column, Offset: condEnd.Offset,
		},
		Kind:        mutation.ConditionalExprTrue,
		Status:      mutation.Runnable,
		Original:    original,
		Replacement: "true",
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IfStmt.Cond",
				Action:      mutation.ActionReplaceWithTrue,
				TargetIndex: -1,
				StartOffset: condStart.Offset,
				EndOffset:   condEnd.Offset,
			},
		},
	}

	if !yield(descTrue) {
		return
	}

	// Replace with false.
	descFalse := mutation.Descriptor{
		File:    ctx.FilePath,
		PkgPath: ctx.PkgPath,
		StartPos: mutation.Position{
			Line: condStart.Line, Column: condStart.Column, Offset: condStart.Offset,
		},
		EndPos: mutation.Position{
			Line: condEnd.Line, Column: condEnd.Column, Offset: condEnd.Offset,
		},
		Kind:        mutation.ConditionalExprFalse,
		Status:      mutation.Runnable,
		Original:    original,
		Replacement: "false",
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IfStmt.Cond",
				Action:      mutation.ActionReplaceWithFalse,
				TargetIndex: -1,
				StartOffset: condStart.Offset,
				EndOffset:   condEnd.Offset,
			},
		},
	}

	yield(descFalse)
}

func discoverForCond(ctx *strategy.DiscoveryContext, n *ast.ForStmt, yield func(mutation.Descriptor) bool) {
	if n.Cond == nil {
		return
	}

	if isBoolLiteral(n.Cond) {
		return
	}

	condStart := ctx.Fset.Position(n.Cond.Pos())
	condEnd := ctx.Fset.Position(n.Cond.End())
	original := sourceText(ctx.Src, condStart.Offset, condEnd.Offset)

	// Only replace with false (true would create infinite loop).
	desc := mutation.Descriptor{
		File:    ctx.FilePath,
		PkgPath: ctx.PkgPath,
		StartPos: mutation.Position{
			Line: condStart.Line, Column: condStart.Column, Offset: condStart.Offset,
		},
		EndPos: mutation.Position{
			Line: condEnd.Line, Column: condEnd.Column, Offset: condEnd.Offset,
		},
		Kind:        mutation.ConditionalExprFalse,
		Status:      mutation.Runnable,
		Original:    original,
		Replacement: "false",
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "ForStmt.Cond",
				Action:      mutation.ActionReplaceWithFalse,
				TargetIndex: -1,
				StartOffset: condStart.Offset,
				EndOffset:   condEnd.Offset,
			},
		},
	}

	yield(desc)
}

func isBoolLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "true" || ident.Name == "false"
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<expr>"
}
