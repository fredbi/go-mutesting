package statement

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&removeStrategy{})
}

// removeStrategy replaces removable statements with noops.
// Ported from go-mutesting's statement/remove mutator.
type removeStrategy struct{}

func (s *removeStrategy) Name() string { return "statement/remove" }
func (s *removeStrategy) NodeTypes() []string {
	return []string{"*ast.BlockStmt", "*ast.CaseClause"}
}

func (s *removeStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		var stmts []ast.Stmt

		switch n := node.(type) {
		case *ast.BlockStmt:
			stmts = n.List
		case *ast.CaseClause:
			stmts = n.Body
		default:
			return
		}

		for i, stmt := range stmts {
			if !isRemovable(stmt) {
				continue
			}

			stmtStart := ctx.Fset.Position(stmt.Pos())
			stmtEnd := ctx.Fset.Position(stmt.End())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: stmtStart.Line, Column: stmtStart.Column, Offset: stmtStart.Offset},
				EndPos:   mutation.Position{Line: stmtEnd.Line, Column: stmtEnd.Column, Offset: stmtEnd.Offset},
				Kind:     mutation.StatementRemove,
				Status:   mutation.Runnable,
				Original:    describeStmt(stmt),
				Replacement: "noop",
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:    nodeTypeName(node),
						Action:      mutation.ActionRemoveStatement,
						TargetIndex: i,
						StartOffset: stmtStart.Offset,
						EndOffset:   stmtEnd.Offset,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}

// isRemovable checks if a statement can be safely removed.
// Mirrors go-mutesting's checkRemoveStatement logic:
// - AssignStmt only if not a short variable declaration (:=)
// - ExprStmt and IncDecStmt are always removable
func isRemovable(stmt ast.Stmt) bool {
	switch n := stmt.(type) {
	case *ast.AssignStmt:
		return n.Tok != token.DEFINE
	case *ast.ExprStmt, *ast.IncDecStmt:
		return true
	}
	return false
}

func nodeTypeName(node ast.Node) string {
	switch node.(type) {
	case *ast.BlockStmt:
		return "BlockStmt"
	case *ast.CaseClause:
		return "CaseClause"
	default:
		return "Unknown"
	}
}

func describeStmt(stmt ast.Stmt) string {
	switch stmt.(type) {
	case *ast.AssignStmt:
		return "assignment"
	case *ast.ExprStmt:
		return "expression statement"
	case *ast.IncDecStmt:
		return "inc/dec statement"
	default:
		return "statement"
	}
}
