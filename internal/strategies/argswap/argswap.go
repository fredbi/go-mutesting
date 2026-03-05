package argswap

import (
	"fmt"
	"go/ast"
	"go/types"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&argSwapStrategy{})
}

// argSwapStrategy swaps adjacent function call arguments of the same type.
//
// For example, f(a, b, flag) where a and b are both int produces:
//
//	f(b, a, flag)
//
// Requires types.Info.Types to be populated (the type checker must be configured
// with a non-nil Types map). If type info is unavailable, no mutations are generated.
type argSwapStrategy struct{}

func (s *argSwapStrategy) Name() string        { return "argswap/swap_arguments" }
func (s *argSwapStrategy) NodeTypes() []string  { return []string{"*ast.CallExpr"} }

func (s *argSwapStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return
		}

		// Skip variadic expansion calls: f(slice...)
		if call.Ellipsis.IsValid() {
			return
		}

		// Need type info to compare argument types.
		if ctx.Info == nil || ctx.Info.Types == nil {
			return
		}

		for i := 0; i < len(call.Args)-1; i++ {
			argA := call.Args[i]
			argB := call.Args[i+1]

			tvA, okA := ctx.Info.Types[argA]
			tvB, okB := ctx.Info.Types[argB]
			if !okA || !okB {
				continue
			}

			if !types.Identical(tvA.Type, tvB.Type) {
				continue
			}

			// Skip if both arguments have identical source text (equivalent mutant).
			srcA := sourceText(ctx.Src, ctx.Fset.Position(argA.Pos()).Offset, ctx.Fset.Position(argA.End()).Offset)
			srcB := sourceText(ctx.Src, ctx.Fset.Position(argB.Pos()).Offset, ctx.Fset.Position(argB.End()).Offset)
			if srcA == srcB {
				continue
			}

			callStart := ctx.Fset.Position(call.Lparen)
			callEnd := ctx.Fset.Position(call.Rparen)

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: callStart.Line, Column: callStart.Column, Offset: callStart.Offset,
				},
				EndPos: mutation.Position{
					Line: callEnd.Line, Column: callEnd.Column, Offset: callEnd.Offset,
				},
				Kind:        mutation.ArgSwap,
				Status:      mutation.Runnable,
				Original:    fmt.Sprintf("%s, %s", srcA, srcB),
				Replacement: fmt.Sprintf("%s, %s", srcB, srcA),
				Apply: mutation.ApplySpec{
					Structural: &mutation.StructuralSpec{
						NodeType:     "CallExpr",
						Action:       mutation.ActionSwapCallArgs,
						TargetIndex:  i,
						TargetIndex2: i + 1,
						StartOffset:  callStart.Offset,
						EndOffset:    callEnd.Offset,
					},
				},
			}

			if !yield(desc) {
				return
			}
		}
	}
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<expr>"
}
