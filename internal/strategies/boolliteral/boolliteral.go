package boolliteral

import (
	"go/ast"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&boolLiteralStrategy{})
}

type boolLiteralStrategy struct{}

func (s *boolLiteralStrategy) Name() string       { return "bool_literal" }
func (s *boolLiteralStrategy) NodeTypes() []string { return []string{"*ast.Ident"} }

func (s *boolLiteralStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		ident, ok := node.(*ast.Ident)
		if !ok {
			return
		}

		var kind mutation.Kind
		var original, replacement string

		switch ident.Name {
		case "true":
			kind = mutation.BoolLitTrueToFalse
			original = "true"
			replacement = "false"
		case "false":
			kind = mutation.BoolLitFalseToTrue
			original = "false"
			replacement = "true"
		default:
			return
		}

		// Verify it's actually a boolean literal, not a user-defined identifier.
		// A boolean literal has Obj == nil (it's a predeclared identifier from the universe scope).
		if ident.Obj != nil {
			return
		}

		pos := ctx.Fset.Position(ident.Pos())
		endOffset := pos.Offset + len(ident.Name)

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: pos.Line, Column: pos.Column, Offset: pos.Offset,
			},
			EndPos: mutation.Position{
				Line: pos.Line, Column: pos.Column + len(ident.Name), Offset: endOffset,
			},
			Kind:        kind,
			Status:      mutation.Runnable,
			Original:    original,
			Replacement: replacement,
			Apply: mutation.ApplySpec{
				TokenSwap: &mutation.TokenSwapSpec{
					OriginalToken:    original,
					ReplacementToken: replacement,
					StartOffset:      pos.Offset,
					EndOffset:        endOffset,
				},
			},
		}

		yield(desc)
	}
}
