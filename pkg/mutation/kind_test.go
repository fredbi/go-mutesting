package mutation

import "testing"

func TestKindCategoryAndName(t *testing.T) {
	tests := []struct {
		kind     Kind
		category string
		name     string
	}{
		{"arithmetic/add_to_sub", "arithmetic", "add_to_sub"},
		{"branch/empty_if", "branch", "empty_if"},
		{"standalone", "standalone", ""},
	}

	for _, tt := range tests {
		if got := tt.kind.Category(); got != tt.category {
			t.Errorf("Kind(%q).Category() = %q, want %q", tt.kind, got, tt.category)
		}
		if got := tt.kind.Name(); got != tt.name {
			t.Errorf("Kind(%q).Name() = %q, want %q", tt.kind, got, tt.name)
		}
	}
}
