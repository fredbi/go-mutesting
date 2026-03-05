package returnval

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&nilErrorStrategy{})
	strategy.Register(&zeroValueStrategy{})
	strategy.Register(&negateBoolStrategy{})
}

// nilErrorStrategy: for func() (..., error), replace "return ..., err" with "return ..., nil".
type nilErrorStrategy struct{}

func (s *nilErrorStrategy) Name() string       { return "return/nil_error" }
func (s *nilErrorStrategy) NodeTypes() []string { return []string{"*ast.ReturnStmt"} }

func (s *nilErrorStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return
		}

		// Find the enclosing function's result types.
		results := enclosingResultList(ctx, ret)
		if results == nil || results.NumFields() == 0 {
			return
		}

		// We need the return to have explicit results (not a naked return).
		if len(ret.Results) == 0 {
			return
		}

		resultTypes := flattenFieldTypes(results)

		// Find the last result that is the error type.
		errIdx := -1
		for i := len(resultTypes) - 1; i >= 0; i-- {
			if isErrorType(resultTypes[i]) {
				errIdx = i
				break
			}
		}

		if errIdx < 0 {
			return
		}

		// Don't mutate if the error value is already nil.
		if isNilIdent(ret.Results[errIdx]) {
			return
		}

		retStart := ctx.Fset.Position(ret.Pos())
		retEnd := ctx.Fset.Position(ret.End())

		// Build ReturnMeta: empty = keep, "nil" = replace.
		meta := make([]string, len(ret.Results))
		meta[errIdx] = "nil"

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: retStart.Line, Column: retStart.Column, Offset: retStart.Offset,
			},
			EndPos: mutation.Position{
				Line: retEnd.Line, Column: retEnd.Column, Offset: retEnd.Offset,
			},
			Kind:        mutation.ReturnNilError,
			Status:      mutation.Runnable,
			Original:    exprString(ctx, ret.Results[errIdx]),
			Replacement: "nil",
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "ReturnStmt",
					Action:      mutation.ActionReturnZero,
					TargetIndex: errIdx,
					StartOffset: retStart.Offset,
					EndOffset:   retEnd.Offset,
					ReturnMeta:  meta,
				},
			},
		}

		yield(desc)
	}
}

// zeroValueStrategy: for func() (T, ..., error), replace "return val, ..., err"
// with "return <zero>, ..., err" — zeroing non-error return values.
type zeroValueStrategy struct{}

func (s *zeroValueStrategy) Name() string       { return "return/zero_value" }
func (s *zeroValueStrategy) NodeTypes() []string { return []string{"*ast.ReturnStmt"} }

func (s *zeroValueStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return
		}

		results := enclosingResultList(ctx, ret)
		if results == nil || results.NumFields() == 0 {
			return
		}

		if len(ret.Results) == 0 {
			return
		}

		resultTypes := flattenFieldTypes(results)
		if len(resultTypes) != len(ret.Results) {
			return
		}

		// For each non-error result, generate a mutation that zeros it.
		for i, typ := range resultTypes {
			if isErrorType(typ) {
				continue
			}

			zero := zeroLiteral(typ)
			if zero == "" {
				continue
			}

			// Don't mutate if already zero.
			if exprString(ctx, ret.Results[i]) == zero {
				continue
			}

			retStart := ctx.Fset.Position(ret.Pos())
			retEnd := ctx.Fset.Position(ret.End())

			meta := make([]string, len(ret.Results))
			meta[i] = zero

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: retStart.Line, Column: retStart.Column, Offset: retStart.Offset,
				},
				EndPos: mutation.Position{
					Line: retEnd.Line, Column: retEnd.Column, Offset: retEnd.Offset,
				},
				Kind:        mutation.ReturnZeroValue,
				Status:      mutation.Runnable,
				Original:    exprString(ctx, ret.Results[i]),
				Replacement: zero,
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:    "ReturnStmt",
						Action:      mutation.ActionReturnZero,
						TargetIndex: i,
						StartOffset: retStart.Offset,
						EndOffset:   retEnd.Offset,
						ReturnMeta:  meta,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}

// negateBoolStrategy: for func() (bool, ...), replace "return true/boolExpr, ..."
// with "return !boolExpr, ...".
type negateBoolStrategy struct{}

func (s *negateBoolStrategy) Name() string       { return "return/negate_bool" }
func (s *negateBoolStrategy) NodeTypes() []string { return []string{"*ast.ReturnStmt"} }

func (s *negateBoolStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return
		}

		results := enclosingResultList(ctx, ret)
		if results == nil || results.NumFields() == 0 {
			return
		}

		if len(ret.Results) == 0 {
			return
		}

		resultTypes := flattenFieldTypes(results)
		if len(resultTypes) != len(ret.Results) {
			return
		}

		for i, typ := range resultTypes {
			if !isBoolType(typ) {
				continue
			}

			// Don't negate an already negated expression (avoid double negation).
			if isNegation(ret.Results[i]) {
				continue
			}

			retStart := ctx.Fset.Position(ret.Pos())
			retEnd := ctx.Fset.Position(ret.End())

			original := exprString(ctx, ret.Results[i])
			replacement := "!" + original

			meta := make([]string, len(ret.Results))
			meta[i] = "!" // special marker: negate

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: retStart.Line, Column: retStart.Column, Offset: retStart.Offset,
				},
				EndPos: mutation.Position{
					Line: retEnd.Line, Column: retEnd.Column, Offset: retEnd.Offset,
				},
				Kind:        mutation.ReturnNegateBool,
				Status:      mutation.Runnable,
				Original:    original,
				Replacement: replacement,
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:    "ReturnStmt",
						Action:      mutation.ActionNegateBoolReturn,
						TargetIndex: i,
						StartOffset: retStart.Offset,
						EndOffset:   retEnd.Offset,
						ReturnMeta:  meta,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}

// --- helpers ---

// enclosingResultList finds the result FieldList of the function enclosing the given return statement.
func enclosingResultList(ctx *strategy.DiscoveryContext, ret *ast.ReturnStmt) *ast.FieldList {
	retPos := ret.Pos()
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

		if n.Pos() <= retPos && retPos < n.End() {
			result = funcType.Results
			return true // keep going to find innermost
		}
		return true
	})

	return result
}

// flattenFieldTypes returns the ordered list of types from a FieldList,
// expanding fields with multiple names (e.g. "a, b int" → [int, int]).
func flattenFieldTypes(fl *ast.FieldList) []types.Type {
	if fl == nil {
		return nil
	}

	var result []types.Type
	for _, field := range fl.List {
		typ := fieldType(field)
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

// fieldType resolves the types.Type for an ast.Field.
// It works by looking at the field's type expression.
func fieldType(field *ast.Field) types.Type {
	return exprToType(field.Type)
}

// exprToType converts common type expressions to types.Type.
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
		// Named type — treat as pointer-like (nilable).
		if e.Obj != nil {
			return nil // can't resolve without full type info
		}
		return nil
	case *ast.StarExpr:
		return types.NewPointer(types.Typ[types.Int]) // sentinel: pointer type
	case *ast.ArrayType:
		if e.Len == nil {
			return types.NewSlice(types.Typ[types.Int]) // sentinel: slice type
		}
		return nil
	case *ast.MapType:
		return types.NewMap(types.Typ[types.String], types.Typ[types.Int]) // sentinel
	case *ast.InterfaceType:
		return types.NewInterfaceType(nil, nil)
	case *ast.SelectorExpr:
		return nil // can't resolve cross-package without full type info
	case *ast.ChanType:
		return nil
	case *ast.FuncType:
		return nil
	}
	return nil
}

func isErrorType(typ types.Type) bool {
	if typ == nil {
		return false
	}
	return typ.String() == "error"
}

func isBoolType(typ types.Type) bool {
	if typ == nil {
		return false
	}
	basic, ok := typ.(*types.Basic)
	return ok && basic.Kind() == types.Bool
}

func isNilIdent(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}

func isNegation(expr ast.Expr) bool {
	unary, ok := expr.(*ast.UnaryExpr)
	return ok && unary.Op == token.NOT
}

// zeroLiteral returns the Go zero-value literal for a type, or "" if unknown.
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

// exprString returns the source text of an expression.
func exprString(ctx *strategy.DiscoveryContext, expr ast.Expr) string {
	start := ctx.Fset.Position(expr.Pos())
	end := ctx.Fset.Position(expr.End())

	if start.Offset >= 0 && end.Offset <= len(ctx.Src) && start.Offset < end.Offset {
		return string(ctx.Src[start.Offset:end.Offset])
	}

	// Fallback for synthetic nodes.
	return fmt.Sprintf("expr@%d:%d", start.Line, start.Column)
}
