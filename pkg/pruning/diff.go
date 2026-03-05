package pruning

import (
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// DiffData holds parsed diff information.
// Maps file path -> set of changed line numbers.
type DiffData struct {
	ChangedLines map[string]map[int]bool
}

// DiffFilter marks mutations outside changed lines as Skipped.
type DiffFilter struct {
	Diff *DiffData
}

func (f *DiffFilter) Name() string { return "diff" }

func (f *DiffFilter) Apply(source iter.Seq[mutation.Descriptor]) iter.Seq[mutation.Descriptor] {
	if f.Diff == nil {
		return source
	}

	return func(yield func(mutation.Descriptor) bool) {
		for desc := range source {
			if !f.inDiff(desc) {
				desc.Status = mutation.Skipped
			}
			if !yield(desc) {
				return
			}
		}
	}
}

func (f *DiffFilter) inDiff(desc mutation.Descriptor) bool {
	lines, ok := f.Diff.ChangedLines[desc.File]
	if !ok {
		return false
	}
	// Check if any line in the mutation range is in the diff.
	for l := desc.StartPos.Line; l <= desc.EndPos.Line; l++ {
		if lines[l] {
			return true
		}
	}
	return false
}
