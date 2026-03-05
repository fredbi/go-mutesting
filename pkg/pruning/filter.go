package pruning

import (
	"iter"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// Filter transforms a mutation stream, potentially setting Status on filtered-out descriptors.
type Filter interface {
	Name() string
	Apply(iter.Seq[mutation.Descriptor]) iter.Seq[mutation.Descriptor]
}

// Chain composes filters: each wraps the iterator from the previous one.
func Chain(source iter.Seq[mutation.Descriptor], filters ...Filter) iter.Seq[mutation.Descriptor] {
	for _, f := range filters {
		source = f.Apply(source)
	}
	return source
}
