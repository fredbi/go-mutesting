package slicebound

import (
	"fmt"
	"go/ast"
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&sliceBoundStrategy{})
}

// sliceBoundStrategy introduces off-by-one errors in index and slice expressions.
//
//	s[i]    → s[i+1], s[i-1]
//	s[lo:hi] → s[lo:hi+1], s[lo+1:hi]
//
// These mutations typically cause panics (index out of range, slice bounds out of range).
type sliceBoundStrategy struct{}

func (s *sliceBoundStrategy) Name() string { return "slicebound" }
func (s *sliceBoundStrategy) NodeTypes() []string {
	return []string{"*ast.IndexExpr", "*ast.SliceExpr"}
}

func (s *sliceBoundStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		switch n := node.(type) {
		case *ast.IndexExpr:
			discoverIndex(ctx, n, yield)
		case *ast.SliceExpr:
			discoverSlice(ctx, n, yield)
		}
	}
}

func discoverIndex(ctx *strategy.DiscoveryContext, ie *ast.IndexExpr, yield func(mutation.Descriptor) bool) {
	idxStart := ctx.Fset.Position(ie.Index.Pos())
	idxEnd := ctx.Fset.Position(ie.Index.End())
	original := sourceText(ctx.Src, idxStart.Offset, idxEnd.Offset)

	// s[i] → s[i+1]
	descUp := mutation.Descriptor{
		File:    ctx.FilePath,
		PkgPath: ctx.PkgPath,
		StartPos: mutation.Position{
			Line: idxStart.Line, Column: idxStart.Column, Offset: idxStart.Offset,
		},
		EndPos: mutation.Position{
			Line: idxEnd.Line, Column: idxEnd.Column, Offset: idxEnd.Offset,
		},
		Kind:        mutation.SliceBoundIndexUp,
		Status:      mutation.Runnable,
		Original:    original,
		Replacement: fmt.Sprintf("%s + 1", original),
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IndexExpr.Index",
				Action:      mutation.ActionOffsetExpr,
				TargetIndex: 1, // +1
				StartOffset: idxStart.Offset,
				EndOffset:   idxEnd.Offset,
			},
		},
	}

	if !yield(descUp) {
		return
	}

	// s[i] → s[i-1]
	descDown := mutation.Descriptor{
		File:    ctx.FilePath,
		PkgPath: ctx.PkgPath,
		StartPos: mutation.Position{
			Line: idxStart.Line, Column: idxStart.Column, Offset: idxStart.Offset,
		},
		EndPos: mutation.Position{
			Line: idxEnd.Line, Column: idxEnd.Column, Offset: idxEnd.Offset,
		},
		Kind:        mutation.SliceBoundIndexDown,
		Status:      mutation.Runnable,
		Original:    original,
		Replacement: fmt.Sprintf("%s - 1", original),
		Apply: mutation.ApplySpec{
			Structural: &mutation.StructuralSpec{
				NodeType:    "IndexExpr.Index",
				Action:      mutation.ActionOffsetExpr,
				TargetIndex: -1, // -1
				StartOffset: idxStart.Offset,
				EndOffset:   idxEnd.Offset,
			},
		},
	}

	yield(descDown)
}

func discoverSlice(ctx *strategy.DiscoveryContext, se *ast.SliceExpr, yield func(mutation.Descriptor) bool) {
	// s[lo:hi] → s[lo:hi+1]
	if se.High != nil {
		hiStart := ctx.Fset.Position(se.High.Pos())
		hiEnd := ctx.Fset.Position(se.High.End())
		original := sourceText(ctx.Src, hiStart.Offset, hiEnd.Offset)

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: hiStart.Line, Column: hiStart.Column, Offset: hiStart.Offset,
			},
			EndPos: mutation.Position{
				Line: hiEnd.Line, Column: hiEnd.Column, Offset: hiEnd.Offset,
			},
			Kind:        mutation.SliceBoundHighUp,
			Status:      mutation.Runnable,
			Original:    original,
			Replacement: fmt.Sprintf("%s + 1", original),
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "SliceExpr.High",
					Action:      mutation.ActionOffsetExpr,
					TargetIndex: 1,
					StartOffset: hiStart.Offset,
					EndOffset:   hiEnd.Offset,
				},
			},
		}

		if !yield(desc) {
			return
		}
	}

	// s[lo:hi] → s[lo+1:hi]
	if se.Low != nil {
		loStart := ctx.Fset.Position(se.Low.Pos())
		loEnd := ctx.Fset.Position(se.Low.End())
		original := sourceText(ctx.Src, loStart.Offset, loEnd.Offset)

		desc := mutation.Descriptor{
			File:    ctx.FilePath,
			PkgPath: ctx.PkgPath,
			StartPos: mutation.Position{
				Line: loStart.Line, Column: loStart.Column, Offset: loStart.Offset,
			},
			EndPos: mutation.Position{
				Line: loEnd.Line, Column: loEnd.Column, Offset: loEnd.Offset,
			},
			Kind:        mutation.SliceBoundLowUp,
			Status:      mutation.Runnable,
			Original:    original,
			Replacement: fmt.Sprintf("%s + 1", original),
			Apply: mutation.ApplySpec{
				Structural: &mutation.StructuralSpec{
					NodeType:    "SliceExpr.Low",
					Action:      mutation.ActionOffsetExpr,
					TargetIndex: 1,
					StartOffset: loStart.Offset,
					EndOffset:   loEnd.Offset,
				},
			},
		}

		yield(desc)
	}
}

func sourceText(src []byte, start, end int) string {
	if start >= 0 && end <= len(src) && start < end {
		return string(src[start:end])
	}
	return "<expr>"
}
