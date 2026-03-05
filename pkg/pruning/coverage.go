package pruning

import (
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// CoverageData holds parsed coverage information.
// Maps file path -> list of covered line ranges.
type CoverageData struct {
	// CoveredLines maps absolute file path to a set of covered line numbers.
	CoveredLines map[string]map[int]bool
}

// CoverageFilter marks mutations in uncovered code as NotCovered.
type CoverageFilter struct {
	Coverage *CoverageData
}

func (f *CoverageFilter) Name() string { return "coverage" }

func (f *CoverageFilter) Apply(source iter.Seq[mutation.Descriptor]) iter.Seq[mutation.Descriptor] {
	if f.Coverage == nil {
		return source
	}

	return func(yield func(mutation.Descriptor) bool) {
		for desc := range source {
			if !f.isCovered(desc) {
				desc.Status = mutation.NotCovered
			}
			if !yield(desc) {
				return
			}
		}
	}
}

func (f *CoverageFilter) isCovered(desc mutation.Descriptor) bool {
	lines, ok := f.Coverage.CoveredLines[desc.File]
	if !ok {
		return false
	}
	return lines[desc.StartPos.Line]
}
