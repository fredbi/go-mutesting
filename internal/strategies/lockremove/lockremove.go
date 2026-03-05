package lockremove

import (
	"go/ast"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&lockRemoveStrategy{})
}

// lockRemoveStrategy removes mutex Lock/Unlock/RLock/RUnlock calls.
//
//   - Removing Lock causes the subsequent Unlock to panic (unlock of unlocked mutex).
//   - Removing Unlock causes the next Lock to deadlock (HANG).
//   - Same for RLock/RUnlock.
//
// Also targets defer statements: defer mu.Unlock() is a common Go pattern.
// Method identification is by name (Lock, Unlock, RLock, RUnlock) with zero arguments.
type lockRemoveStrategy struct{}

func (s *lockRemoveStrategy) Name() string { return "lockremove" }
func (s *lockRemoveStrategy) NodeTypes() []string {
	return []string{"*ast.BlockStmt", "*ast.CaseClause"}
}

func (s *lockRemoveStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
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
			kind, methodName := matchLockCall(stmt)
			if kind == "" {
				continue
			}

			stmtStart := ctx.Fset.Position(stmt.Pos())
			stmtEnd := ctx.Fset.Position(stmt.End())

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: stmtStart.Line, Column: stmtStart.Column, Offset: stmtStart.Offset,
				},
				EndPos: mutation.Position{
					Line: stmtEnd.Line, Column: stmtEnd.Column, Offset: stmtEnd.Offset,
				},
				Kind:        kind,
				Status:      mutation.Runnable,
				Original:    sourceText(ctx.Src, stmtStart.Offset, stmtEnd.Offset),
				Replacement: "noop (remove " + methodName + ")",
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:    "BlockStmt",
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

// matchLockCall checks if a statement is a Lock/Unlock/RLock/RUnlock call
// (either direct or via defer).
func matchLockCall(stmt ast.Stmt) (mutation.Kind, string) {
	var call *ast.CallExpr
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		var ok bool
		call, ok = s.X.(*ast.CallExpr)
		if !ok {
			return "", ""
		}
	case *ast.DeferStmt:
		call = s.Call
	default:
		return "", ""
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}

	// Lock/Unlock/RLock/RUnlock take no arguments.
	if len(call.Args) != 0 {
		return "", ""
	}

	switch sel.Sel.Name {
	case "Lock":
		return mutation.LockRemoveLock, "Lock"
	case "Unlock":
		return mutation.LockRemoveUnlock, "Unlock"
	case "RLock":
		return mutation.LockRemoveRLock, "RLock"
	case "RUnlock":
		return mutation.LockRemoveRUnlock, "RUnlock"
	}

	return "", ""
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<stmt>"
}
