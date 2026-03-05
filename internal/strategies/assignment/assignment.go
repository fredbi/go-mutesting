package assignment

import (
	"go/ast"
	"go/token"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&invertStrategy{})
	strategy.Register(&removeSelfStrategy{})
	strategy.Register(&bitwiseAssignStrategy{})
}

// Invert assignment mutations: += <-> -=, *= <-> /=
var invertMutations = []struct {
	from token.Token
	to   token.Token
	kind mutation.Kind
}{
	{token.ADD_ASSIGN, token.SUB_ASSIGN, mutation.AssignmentInvertAddAssign},
	{token.SUB_ASSIGN, token.ADD_ASSIGN, mutation.AssignmentInvertSubAssign},
	{token.MUL_ASSIGN, token.QUO_ASSIGN, mutation.AssignmentInvertMulAssign},
	{token.QUO_ASSIGN, token.MUL_ASSIGN, mutation.AssignmentInvertDivAssign},
}

type invertStrategy struct{}

func (s *invertStrategy) Name() string        { return "assignment/invert" }
func (s *invertStrategy) NodeTypes() []string  { return []string{"*ast.AssignStmt"} }

func (s *invertStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.AssignStmt)
		if !ok {
			return
		}

		for _, m := range invertMutations {
			if n.Tok != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.TokPos)
			endOffset := pos.Offset + len(n.Tok.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Tok.String()), Offset: endOffset},
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

// Remove-self mutations: +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>= -> =
var removeSelfMutations = []struct {
	from token.Token
	kind mutation.Kind
}{
	{token.ADD_ASSIGN, mutation.AssignmentRemoveAdd},
	{token.SUB_ASSIGN, mutation.AssignmentRemoveSub},
	{token.MUL_ASSIGN, mutation.AssignmentRemoveMul},
	{token.QUO_ASSIGN, mutation.AssignmentRemoveDiv},
	{token.REM_ASSIGN, mutation.AssignmentRemoveRem},
	{token.AND_ASSIGN, mutation.AssignmentRemoveAndAssign},
	{token.OR_ASSIGN, mutation.AssignmentRemoveOrAssign},
	{token.XOR_ASSIGN, mutation.AssignmentRemoveXorAssign},
	{token.SHL_ASSIGN, mutation.AssignmentRemoveShlAssign},
	{token.SHR_ASSIGN, mutation.AssignmentRemoveShrAssign},
}

type removeSelfStrategy struct{}

func (s *removeSelfStrategy) Name() string        { return "assignment/remove_self" }
func (s *removeSelfStrategy) NodeTypes() []string  { return []string{"*ast.AssignStmt"} }

func (s *removeSelfStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.AssignStmt)
		if !ok {
			return
		}

		for _, m := range removeSelfMutations {
			if n.Tok != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.TokPos)
			endOffset := pos.Offset + len(n.Tok.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Tok.String()), Offset: endOffset},
				Kind:     m.kind,
				Status:   mutation.Runnable,
				Original:    m.from.String(),
				Replacement: token.ASSIGN.String(),
				Apply: mutation.ApplySpec{
					TokenSwap: &mutation.TokenSwapSpec{
						OriginalToken:    m.from.String(),
						ReplacementToken: token.ASSIGN.String(),
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

// Bitwise assignment mutations: &= <-> |=, <<= <-> >>=
var bitwiseAssignMutations = []struct {
	from token.Token
	to   token.Token
	kind mutation.Kind
}{
	{token.AND_ASSIGN, token.OR_ASSIGN, mutation.BitwiseAssignAndToOr},
	{token.OR_ASSIGN, token.AND_ASSIGN, mutation.BitwiseAssignOrToAnd},
	{token.SHL_ASSIGN, token.SHR_ASSIGN, mutation.BitwiseAssignShlToShr},
	{token.SHR_ASSIGN, token.SHL_ASSIGN, mutation.BitwiseAssignShrToShl},
}

type bitwiseAssignStrategy struct{}

func (s *bitwiseAssignStrategy) Name() string        { return "bitwise_assign" }
func (s *bitwiseAssignStrategy) NodeTypes() []string  { return []string{"*ast.AssignStmt"} }

func (s *bitwiseAssignStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.AssignStmt)
		if !ok {
			return
		}

		for _, m := range bitwiseAssignMutations {
			if n.Tok != m.from {
				continue
			}

			pos := ctx.Fset.Position(n.TokPos)
			endOffset := pos.Offset + len(n.Tok.String())

			desc := mutation.Descriptor{
				File:     ctx.FilePath,
				PkgPath:  ctx.PkgPath,
				StartPos: mutation.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
				EndPos:   mutation.Position{Line: pos.Line, Column: pos.Column + len(n.Tok.String()), Offset: endOffset},
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
