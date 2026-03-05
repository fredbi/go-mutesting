package walker

import (
	"go/ast"
	"go/token"
)

// funcInterval maps a byte offset range to a function name.
type funcInterval struct {
	start int // byte offset
	end   int // byte offset
	name  string
}

// funcNameResolver resolves the enclosing function name for a given position.
type funcNameResolver struct {
	intervals []funcInterval
}

// buildFuncNameResolver builds a function-name interval map from the AST file.
func buildFuncNameResolver(fset *token.FileSet, file *ast.File) *funcNameResolver {
	r := &funcNameResolver{}

	ast.Inspect(file, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		name := fd.Name.Name
		if fd.Recv != nil && len(fd.Recv.List) > 0 {
			name = receiverTypeName(fd.Recv.List[0].Type) + "." + name
		}

		startPos := fset.Position(fd.Pos())
		endPos := fset.Position(fd.End())

		r.intervals = append(r.intervals, funcInterval{
			start: startPos.Offset,
			end:   endPos.Offset,
			name:  name,
		})

		return false // don't recurse into nested function literals
	})

	return r
}

// resolve returns the enclosing function name for the given byte offset, or "".
func (r *funcNameResolver) resolve(offset int) string {
	for _, iv := range r.intervals {
		if offset >= iv.start && offset < iv.end {
			return iv.name
		}
	}
	return ""
}

// receiverTypeName extracts the type name from a receiver expression.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverTypeName(t.X)
	default:
		return ""
	}
}
