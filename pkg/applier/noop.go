package applier

import (
	"go/ast"
	"go/token"
	"go/types"
)

// createNoopOfStatement creates a syntactically safe noop statement out of a given statement.
// Ported from go-mutesting's astutil.CreateNoopOfStatement.
func createNoopOfStatement(pkg *types.Package, info *types.Info, stmt ast.Stmt) ast.Stmt {
	return createNoopOfStatements(pkg, info, []ast.Stmt{stmt})
}

// createNoopOfStatements creates a syntactically safe noop statement out of given statements.
// Ported from go-mutesting's astutil.CreateNoopOfStatements.
func createNoopOfStatements(pkg *types.Package, info *types.Info, stmts []ast.Stmt) ast.Stmt {
	var ids []ast.Expr
	for _, stmt := range stmts {
		ids = append(ids, identifiersInStatement(pkg, info, stmt)...)
	}

	if len(ids) == 0 {
		return &ast.EmptyStmt{
			Semicolon: token.NoPos,
		}
	}

	lhs := make([]ast.Expr, len(ids))
	for i := range ids {
		lhs[i] = ast.NewIdent("_")
	}

	return &ast.AssignStmt{
		Lhs: lhs,
		Rhs: ids,
		Tok: token.ASSIGN,
	}
}

// identifiersInStatement returns all variable identifiers found in a statement.
// Ported from go-mutesting's astutil.IdentifiersInStatement.
func identifiersInStatement(pkg *types.Package, info *types.Info, stmt ast.Stmt) []ast.Expr {
	w := &identifierWalker{
		pkg:  pkg,
		info: info,
	}
	ast.Walk(w, stmt)
	return w.identifiers
}

type identifierWalker struct {
	identifiers []ast.Expr
	pkg         *types.Package
	info        *types.Info
}

func checkForSelectorExpr(node ast.Expr) bool {
	switch n := node.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return checkForSelectorExpr(n.X)
	}
	return false
}

func (w *identifierWalker) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.Ident:
		if n.Name == "_" {
			return nil
		}
		if token.Lookup(n.Name) != token.IDENT {
			return nil
		}
		if obj, ok := w.info.Uses[n]; ok {
			if _, ok := obj.(*types.Var); !ok {
				return nil
			}
		}
		w.identifiers = append(w.identifiers, &ast.Ident{
			Name: n.Name,
		})
		return nil

	case *ast.SelectorExpr:
		if !checkForSelectorExpr(n) {
			return nil
		}

		initialize := false
		if n.Sel != nil {
			if obj, ok := w.info.Uses[n.Sel]; ok {
				t := obj.Type()
				switch t.Underlying().(type) {
				case *types.Array, *types.Map, *types.Slice, *types.Struct:
					initialize = true
				}
			}
		}

		if initialize {
			w.identifiers = append(w.identifiers, &ast.CompositeLit{
				Type: n,
			})
		} else {
			w.identifiers = append(w.identifiers, n)
		}
		return nil
	}

	return w
}
