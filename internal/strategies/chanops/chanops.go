package chanops

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&chanOpsStrategy{})
}

// chanOpsStrategy removes channel operations: close(ch), ch <- val, and <-ch.
//
// All of these mutations are likely to cause hangs:
//   - Removing close(ch) causes range loops over ch to block forever.
//   - Removing ch <- val causes receivers to block forever.
//   - Removing <-ch may cause senders to block (unbuffered channels).
//
// Also targets defer statements: defer close(ch) is a common pattern.
type chanOpsStrategy struct{}

func (s *chanOpsStrategy) Name() string { return "chanops" }
func (s *chanOpsStrategy) NodeTypes() []string {
	return []string{"*ast.BlockStmt", "*ast.CaseClause"}
}

func (s *chanOpsStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
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
			kind, label := matchChanOp(stmt)
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
				Replacement: "noop (remove " + label + ")",
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

// matchChanOp checks if a statement is a channel operation (close, send, or receive).
func matchChanOp(stmt ast.Stmt) (mutation.Kind, string) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		// close(ch)
		if call, ok := s.X.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "close" && ident.Obj == nil {
				return mutation.ChanOpsRemoveClose, "close"
			}
		}
		// <-ch (standalone receive)
		if unary, ok := s.X.(*ast.UnaryExpr); ok && unary.Op == token.ARROW {
			return mutation.ChanOpsRemoveReceive, "receive"
		}

	case *ast.SendStmt:
		// ch <- val
		return mutation.ChanOpsRemoveSend, "send"

	case *ast.DeferStmt:
		// defer close(ch)
		if ident, ok := s.Call.Fun.(*ast.Ident); ok && ident.Name == "close" && ident.Obj == nil {
			return mutation.ChanOpsRemoveClose, "close"
		}
	}

	return "", ""
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<stmt>"
}
