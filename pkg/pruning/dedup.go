package pruning

import (
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// EquivalentMutationFilter deduplicates mutations by their stable ID.
// This is a batch filter: it materializes the stream internally.
type EquivalentMutationFilter struct{}

func (f *EquivalentMutationFilter) Name() string { return "equivalent_mutation" }

func (f *EquivalentMutationFilter) Apply(source iter.Seq[mutation.Descriptor]) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		seen := make(map[string]struct{})
		for desc := range source {
			if _, ok := seen[desc.ID]; ok {
				desc.Status = mutation.Equivalent
			} else {
				seen[desc.ID] = struct{}{}
			}
			if !yield(desc) {
				return
			}
		}
	}
}
