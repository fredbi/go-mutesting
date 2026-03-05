package conditional

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&boundaryStrategy{})
	strategy.Register(&negationStrategy{})
}

// Boundary mutations: < <-> <=, > <-> >=
var boundaryMutations = []struct {
	from token.Token
	to   token.Token
	kind mutation.Kind
}{
	{token.LSS, token.LEQ, mutation.ConditionalBoundaryLessToLessEq},
	{token.LEQ, token.LSS, mutation.ConditionalBoundaryLessEqToLess},
	{token.GTR, token.GEQ, mutation.ConditionalBoundaryGreaterToGrEq},
	{token.GEQ, token.GTR, mutation.ConditionalBoundaryGrEqToGreater},
}

type boundaryStrategy struct{}

func (s *boundaryStrategy) Name() string        { return "conditional/boundary" }
func (s *boundaryStrategy) NodeTypes() []string  { return []string{"*ast.BinaryExpr"} }

func (s *boundaryStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.BinaryExpr)
		if !ok {
			return
		}

		for _, m := range boundaryMutations {
			if n.Op != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.OpPos)
			endOffset := pos.Offset + len(n.Op.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Op.String()), Offset: endOffset},
				Kind:     m.kind,
				Status:   mutation.Runnable,
				Original:    m.from.String(),
				Replacement: m.to.String(),
				Apply: mutation.ApplySpec{
					TokenSwap: &mutation.TokenSwapSpec{
						OriginalToken:    m.from.String(),
						ReplacementToken: m.to.String(),
						StartOffset:      pos.Offset,
						EndOffset:        endOffset,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}

// Negation mutations: == <-> !=, < <-> >=, > <-> <=
var negationMutations = []struct {
	from token.Token
	to   token.Token
	kind mutation.Kind
}{
	{token.EQL, token.NEQ, mutation.ConditionalNegationEqToNeq},
	{token.NEQ, token.EQL, mutation.ConditionalNegationNeqToEq},
	{token.LSS, token.GEQ, mutation.ConditionalNegationLessToGrEq},
	{token.GEQ, token.LSS, mutation.ConditionalNegationGrEqToLess},
	{token.GTR, token.LEQ, mutation.ConditionalNegationGreaterToLessEq},
	{token.LEQ, token.GTR, mutation.ConditionalNegationLessEqToGreater},
}

type negationStrategy struct{}

func (s *negationStrategy) Name() string        { return "conditional/negation" }
func (s *negationStrategy) NodeTypes() []string  { return []string{"*ast.BinaryExpr"} }

func (s *negationStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.BinaryExpr)
		if !ok {
			return
		}

		for _, m := range negationMutations {
			if n.Op != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.OpPos)
			endOffset := pos.Offset + len(n.Op.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Op.String()), Offset: endOffset},
				Kind:     m.kind,
				Status:   mutation.Runnable,
				Original:    m.from.String(),
				Replacement: m.to.String(),
				Apply: mutation.ApplySpec{
					TokenSwap: &mutation.TokenSwapSpec{
						OriginalToken:    m.from.String(),
						ReplacementToken: m.to.String(),
						StartOffset:      pos.Offset,
						EndOffset:        endOffset,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}
