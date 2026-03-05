package pruning

import (
	"iter"
	"regexp"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// FunctionMatchFilter skips mutations outside matched function names.
// If Patterns is empty, all mutations pass through.
type FunctionMatchFilter struct {
	Patterns []*regexp.Regexp
}

func (f *FunctionMatchFilter) Name() string { return "function_match" }

func (f *FunctionMatchFilter) Apply(source iter.Seq[mutation.Descriptor]) iter.Seq[mutation.Descriptor] {
	if len(f.Patterns) == 0 {
		return source
	}

	return func(yield func(mutation.Descriptor) bool) {
		for desc := range source {
			if !f.matches(desc.FuncName) {
				desc.Status = mutation.Skipped
			}
			if !yield(desc) {
				return
			}
		}
	}
}

func (f *FunctionMatchFilter) matches(funcName string) bool {
	for _, p := range f.Patterns {
		if p.MatchString(funcName) {
			return true
		}
	}
	return false
}
