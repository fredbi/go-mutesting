package swapbranch

import (
	"fmt"
	"go/ast"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&swapIfElseStrategy{})
	strategy.Register(&swapCaseStrategy{})
}

// swapIfElseStrategy swaps the if-body with the else-body.
// Only targets if/else (not else-if chains).
type swapIfElseStrategy struct{}

func (s *swapIfElseStrategy) Name() string       { return "branch/swap_if_else" }
func (s *swapIfElseStrategy) NodeTypes() []string { return []string{"*ast.IfStmt"} }

func (s *swapIfElseStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		n, ok := node.(*ast.IfStmt)
		if !ok {
			return
		}

		if n.Else == nil {
			return
		}

		// Only swap when else is a plain block (not else-if).
		elseBlock, ok := n.Else.(*ast.BlockStmt)
		if !ok {
			return
		}

		ifStart := ctx.Fset.Position(n.Body.Lbrace)
		elseEnd := ctx.Fset.Position(elseBlock.Rbrace)

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: ifStart.Line, Column: ifStart.Column, Offset: ifStart.Offset,
			},
			EndPos: mutation.Position{
				Line: elseEnd.Line, Column: elseEnd.Column, Offset: elseEnd.Offset,
			},
			Kind:   mutation.BranchSwapIfElse,
			Status: mutation.Runnable,
			Original: fmt.Sprintf("if body: %d stmts, else body: %d stmts",
				len(n.Body.List), len(elseBlock.List)),
			Replacement: "swap if <-> else bodies",
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "IfStmt.SwapElse",
					Action:      mutation.ActionSwapIfElse,
					TargetIndex: -1,
					StartOffset: ifStart.Offset,
					EndOffset:   elseEnd.Offset + 1,
				},
			},
		}

		yield(desc)
	}
}

// swapCaseStrategy swaps the bodies of adjacent case clauses in a switch.
type swapCaseStrategy struct{}

func (s *swapCaseStrategy) Name() string       { return "branch/swap_case" }
func (s *swapCaseStrategy) NodeTypes() []string { return []string{"*ast.SwitchStmt", "*ast.TypeSwitchStmt"} }

func (s *swapCaseStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		var body *ast.BlockStmt

		switch n := node.(type) {
		case *ast.SwitchStmt:
			body = n.Body
		case *ast.TypeSwitchStmt:
			body = n.Body
		default:
			return
		}

		if body == nil {
			return
		}

		// Collect case clauses with non-empty bodies.
		type caseInfo struct {
			clause *ast.CaseClause
			index  int // index in body.List
		}

		var cases []caseInfo
		for i, stmt := range body.List {
			cc, ok := stmt.(*ast.CaseClause)
			if !ok || len(cc.Body) == 0 {
				continue
			}
			cases = append(cases, caseInfo{clause: cc, index: i})
		}

		// Generate one mutation per adjacent pair.
		for i := 0; i+1 < len(cases); i++ {
			a := cases[i]
			b := cases[i+1]

			aStart := ctx.Fset.Position(a.clause.Colon)
			bEnd := ctx.Fset.Position(b.clause.Body[len(b.clause.Body)-1].End())

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: aStart.Line, Column: aStart.Column, Offset: aStart.Offset,
				},
				EndPos: mutation.Position{
					Line: bEnd.Line, Column: bEnd.Column, Offset: bEnd.Offset,
				},
				Kind:   mutation.BranchSwapCase,
				Status: mutation.Runnable,
				Original: fmt.Sprintf("case[%d]: %d stmts, case[%d]: %d stmts",
					a.index, len(a.clause.Body), b.index, len(b.clause.Body)),
				Replacement: fmt.Sprintf("swap case[%d] <-> case[%d] bodies", a.index, b.index),
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:     "SwitchStmt.SwapCase",
						Action:       mutation.ActionSwapCase,
						TargetIndex:  a.index,
						TargetIndex2: b.index,
						StartOffset:  aStart.Offset,
						EndOffset:    bEnd.Offset,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}
