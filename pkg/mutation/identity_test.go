package mutation

import "testing"

func TestComputeID(t *testing.T) {
	id1 := ComputeID("/path/file.go", "arithmetic/add_to_sub", 10, 5, "+", "-")
	id2 := ComputeID("/path/file.go", "arithmetic/add_to_sub", 10, 5, "+", "-")
	id3 := ComputeID("/path/file.go", "arithmetic/add_to_sub", 11, 5, "+", "-")

	if len(id1) != 32 {
		t.Errorf("ID length = %d, want 32", len(id1))
	}
	if id1 != id2 {
		t.Error("same inputs should produce same ID")
	}
	if id1 == id3 {
		t.Error("different inputs should produce different ID")
	}
}
