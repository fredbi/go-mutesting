package panicremove

import (
	"fmt"
	"go/ast"
	"go/types"
	"iter"
	"strings"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&panicToReturnStrategy{})
}

// panicToReturnStrategy replaces panic() calls with return statements
// returning zero values of the enclosing function's result types.
//
// Only applies when:
// - The panic call is an ExprStmt (standalone statement, not inside an expression)
// - The enclosing function has at least one return value
type panicToReturnStrategy struct{}

func (s *panicToReturnStrategy) Name() string       { return "panic/replace_with_return" }
func (s *panicToReturnStrategy) NodeTypes() []string { return []string{"*ast.ExprStmt"} }

func (s *panicToReturnStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		exprStmt, ok := node.(*ast.ExprStmt)
		if !ok {
			return
		}

		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			return
		}

		fnIdent, ok := call.Fun.(*ast.Ident)
		if !ok || fnIdent.Name != "panic" {
			return
		}

		// Don't match a user-defined "panic" function — only the builtin.
		if fnIdent.Obj != nil {
			return
		}

		// Find enclosing function's result types.
		results := enclosingResultList(ctx, exprStmt)
		if results == nil || results.NumFields() == 0 {
			// In a function with no return values, panic -> return is less interesting
			// but still valid. Generate "return" with no values.
			stmtStart := ctx.Fset.Position(exprStmt.Pos())
			stmtEnd := ctx.Fset.Position(exprStmt.End())

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: stmtStart.Line, Column: stmtStart.Column, Offset: stmtStart.Offset,
				},
				EndPos: mutation.Position{
					Line: stmtEnd.Line, Column: stmtEnd.Column, Offset: stmtEnd.Offset,
				},
				Kind:        mutation.PanicToReturn,
				Status:      mutation.Runnable,
				Original:    sourceText(ctx.Src, stmtStart.Offset, stmtEnd.Offset),
				Replacement: "return",
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:    "ExprStmt.Panic",
						Action:      mutation.ActionReplaceStmtWithReturn,
						TargetIndex: -1,
						StartOffset: stmtStart.Offset,
						EndOffset:   stmtEnd.Offset,
						ReturnMeta:  nil, // no return values
					},
				},
			}

			yield(desc)
			return
		}

		// Build zero values for each result.
		resultTypes := flattenFieldTypes(results)
		meta := make([]string, len(resultTypes))
		zeroStrs := make([]string, len(resultTypes))

		for i, typ := range resultTypes {
			z := zeroLiteral(typ)
			if z == "" {
				// Can't determine zero value — skip this mutation.
				return
			}
			meta[i] = z
			zeroStrs[i] = z
		}

		stmtStart := ctx.Fset.Position(exprStmt.Pos())
		stmtEnd := ctx.Fset.Position(exprStmt.End())

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: stmtStart.Line, Column: stmtStart.Column, Offset: stmtStart.Offset,
			},
			EndPos: mutation.Position{
				Line: stmtEnd.Line, Column: stmtEnd.Column, Offset: stmtEnd.Offset,
			},
			Kind:        mutation.PanicToReturn,
			Status:      mutation.Runnable,
			Original:    sourceText(ctx.Src, stmtStart.Offset, stmtEnd.Offset),
			Replacement: fmt.Sprintf("return %s", strings.Join(zeroStrs, ", ")),
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "ExprStmt.Panic",
					Action:      mutation.ActionReplaceStmtWithReturn,
					TargetIndex: -1,
					StartOffset: stmtStart.Offset,
					EndOffset:   stmtEnd.Offset,
					ReturnMeta:  meta,
				},
			},
		}

		yield(desc)
	}
}

// --- helpers (same approach as returnval) ---

func enclosingResultList(ctx *strategy.DiscoveryContext, target ast.Node) *ast.FieldList {
	targetPos := target.Pos()
	var result *ast.FieldList

	ast.Inspect(ctx.File, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		var funcType *ast.FuncType
		switch fn := n.(type) {
		case *ast.FuncDecl:
			funcType = fn.Type
		case *ast.FuncLit:
			funcType = fn.Type
		default:
			return true
		}

		if n.Pos() <= targetPos && targetPos < n.End() {
			result = funcType.Results
			return true
		}
		return true
	})

	return result
}

func flattenFieldTypes(fl *ast.FieldList) []types.Type {
	if fl == nil {
		return nil
	}
	var result []types.Type
	for _, field := range fl.List {
		typ := exprToType(field.Type)
		if typ == nil {
			continue
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			result = append(result, typ)
		}
	}
	return result
}

func exprToType(expr ast.Expr) types.Type {
	switch e := expr.(type) {
	case *ast.Ident:
		switch e.Name {
		case "bool":
			return types.Typ[types.Bool]
		case "int":
			return types.Typ[types.Int]
		case "int8":
			return types.Typ[types.Int8]
		case "int16":
			return types.Typ[types.Int16]
		case "int32":
			return types.Typ[types.Int32]
		case "int64":
			return types.Typ[types.Int64]
		case "uint":
			return types.Typ[types.Uint]
		case "uint8":
			return types.Typ[types.Uint8]
		case "uint16":
			return types.Typ[types.Uint16]
		case "uint32":
			return types.Typ[types.Uint32]
		case "uint64":
			return types.Typ[types.Uint64]
		case "uintptr":
			return types.Typ[types.Uintptr]
		case "float32":
			return types.Typ[types.Float32]
		case "float64":
			return types.Typ[types.Float64]
		case "complex64":
			return types.Typ[types.Complex64]
		case "complex128":
			return types.Typ[types.Complex128]
		case "string":
			return types.Typ[types.String]
		case "byte":
			return types.Typ[types.Byte]
		case "rune":
			return types.Typ[types.Rune]
		case "error":
			return types.Universe.Lookup("error").Type()
		}
		return nil
	case *ast.StarExpr:
		return types.NewPointer(types.Typ[types.Int])
	case *ast.ArrayType:
		if e.Len == nil {
			return types.NewSlice(types.Typ[types.Int])
		}
		return nil
	case *ast.MapType:
		return types.NewMap(types.Typ[types.String], types.Typ[types.Int])
	case *ast.InterfaceType:
		return types.NewInterfaceType(nil, nil)
	}
	return nil
}

func zeroLiteral(typ types.Type) string {
	if typ == nil {
		return ""
	}
	switch t := typ.Underlying().(type) {
	case *types.Basic:
		switch {
		case t.Info()&types.IsBoolean != 0:
			return "false"
		case t.Info()&types.IsInteger != 0:
			return "0"
		case t.Info()&types.IsFloat != 0:
			return "0"
		case t.Info()&types.IsComplex != 0:
			return "0"
		case t.Info()&types.IsString != 0:
			return `""`
		}
	case *types.Pointer, *types.Slice, *types.Map, *types.Interface, *types.Signature, *types.Chan:
		return "nil"
	}
	return ""
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<stmt>"
}
