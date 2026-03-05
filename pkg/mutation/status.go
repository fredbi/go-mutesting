package mutation

// Status represents the pre-execution status of a mutation descriptor.
type Status int

const (
	// Runnable means the mutation is covered by tests and ready to execute.
	Runnable Status = iota
	// NotCovered means the mutation is in code not covered by tests.
	NotCovered
	// Skipped means the mutation was excluded by diff/filter/config.
	Skipped
	// Equivalent means the mutation is a duplicate of another.
	Equivalent
)

func (s Status) String() string {
	switch s {
	case Runnable:
		return "runnable"
	case NotCovered:
		return "not_covered"
	case Skipped:
		return "skipped"
	case Equivalent:
		return "equivalent"
	default:
		return "unknown"
	}
}
