package nilguard

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&nilGuardStrategy{})
}

// nilGuardStrategy removes nil check guards by replacing the condition
// with a value that bypasses the safety check.
//
//   - if x != nil { use(x) }  →  if true { use(x) }   // nil dereference when x is nil
//   - if x == nil { return }   →  if false { return }   // falls through to use x when nil
type nilGuardStrategy struct{}

func (s *nilGuardStrategy) Name() string        { return "nilguard/remove_nil_check" }
func (s *nilGuardStrategy) NodeTypes() []string  { return []string{"*ast.IfStmt"} }

func (s *nilGuardStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		ifStmt, ok := node.(*ast.IfStmt)
		if !ok {
			return
		}

		be, ok := ifStmt.Cond.(*ast.BinaryExpr)
		if !ok {
			return
		}

		if be.Op != token.NEQ && be.Op != token.EQL {
			return
		}

		if !isNilIdent(be.X) && !isNilIdent(be.Y) {
			return
		}

		// Determine replacement: bypass the nil check.
		// != nil → true  (always enter body, even when value IS nil)
		// == nil → false (never enter guard, fall through to use nil value)
		var replacement string
		var action mutation.StructuralAction
		if be.Op == token.NEQ {
			replacement = "true"
			action = mutation.ActionReplaceWithTrue
		} else {
			replacement = "false"
			action = mutation.ActionReplaceWithFalse
		}

		condStart := ctx.Fset.Position(ifStmt.Cond.Pos())
		condEnd := ctx.Fset.Position(ifStmt.Cond.End())

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: condStart.Line, Column: condStart.Column, Offset: condStart.Offset,
			},
			EndPos: mutation.Position{
				Line: condEnd.Line, Column: condEnd.Column, Offset: condEnd.Offset,
			},
			Kind:        mutation.NilGuardRemove,
			Status:      mutation.Runnable,
			Original:    sourceText(ctx.Src, condStart.Offset, condEnd.Offset),
			Replacement: replacement,
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "IfStmt.Cond",
					Action:      action,
					TargetIndex: -1,
					StartOffset: condStart.Offset,
					EndOffset:   condEnd.Offset,
				},
			},
		}

		yield(desc)
	}
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<expr>"
}
