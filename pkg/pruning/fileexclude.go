package pruning

import (
	"iter"
	"regexp"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// FileExclusionFilter skips mutations in files matching any of the given regex patterns.
type FileExclusionFilter struct {
	Patterns []*regexp.Regexp
}

func (f *FileExclusionFilter) Name() string { return "file_exclusion" }

func (f *FileExclusionFilter) Apply(source iter.Seq[mutation.Descriptor]) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		for desc := range source {
			if f.matches(desc.File) {
				desc.Status = mutation.Skipped
			}
			if !yield(desc) {
				return
			}
		}
	}
}

func (f *FileExclusionFilter) matches(file string) bool {
	for _, p := range f.Patterns {
		if p.MatchString(file) {
			return true
		}
	}
	return false
}
