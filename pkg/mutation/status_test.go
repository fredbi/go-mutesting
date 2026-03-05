package mutation

import "testing"

func TestStatusString(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{Runnable, "runnable"},
		{NotCovered, "not_covered"},
		{Skipped, "skipped"},
		{Equivalent, "equivalent"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}
