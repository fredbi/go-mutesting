package applier

import (
	"fmt"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// Applier applies a single mutation to a file in an isolated workdir.
type Applier interface {
	// Apply applies a single mutation to the file in workdir.
	Apply(desc mutation.Descriptor, workdir string) error

	// Rollback restores the original file in workdir.
	Rollback(desc mutation.Descriptor, workdir string) error
}

// CompositeApplier routes to TokenApplier or StructuralApplier
// based on which ApplySpec field is set.
type CompositeApplier struct {
	Token      *TokenApplier
	Structural *StructuralApplier
}

// NewCompositeApplier creates a CompositeApplier with default sub-appliers.
func NewCompositeApplier() *CompositeApplier {
	return &CompositeApplier{
		Token:      &TokenApplier{},
		Structural: &StructuralApplier{},
	}
}

func (c *CompositeApplier) Apply(desc mutation.Descriptor, workdir string) error {
	a, err := c.route(desc)
	if err != nil {
		return err
	}
	return a.Apply(desc, workdir)
}

func (c *CompositeApplier) Rollback(desc mutation.Descriptor, workdir string) error {
	a, err := c.route(desc)
	if err != nil {
		return err
	}
	return a.Rollback(desc, workdir)
}

func (c *CompositeApplier) route(desc mutation.Descriptor) (Applier, error) {
	switch {
	case desc.Apply.TokenSwap != nil:
		return c.Token, nil
	case desc.Apply.Structural != nil:
		return c.Structural, nil
	default:
		return nil, fmt.Errorf("descriptor %s has no ApplySpec set", desc.ID)
	}
}
