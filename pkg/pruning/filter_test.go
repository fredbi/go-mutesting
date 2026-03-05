package pruning

import (
	"regexp"
	"slices"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

func makeDescriptors() []mutation.Descriptor {
	return []mutation.Descriptor{
		{ID: "aaa", File: "/src/main.go", FuncName: "Foo", StartPos: mutation.Position{Line: 10}, EndPos: mutation.Position{Line: 10}, Kind: "arithmetic/add_to_sub", Status: mutation.Runnable},
		{ID: "bbb", File: "/src/main_test.go", FuncName: "TestFoo", StartPos: mutation.Position{Line: 20}, EndPos: mutation.Position{Line: 20}, Kind: "arithmetic/sub_to_add", Status: mutation.Runnable},
		{ID: "ccc", File: "/src/util.go", FuncName: "Bar", StartPos: mutation.Position{Line: 5}, EndPos: mutation.Position{Line: 5}, Kind: "branch/empty_if", Status: mutation.Runnable},
		{ID: "aaa", File: "/src/main.go", FuncName: "Foo", StartPos: mutation.Position{Line: 10}, EndPos: mutation.Position{Line: 10}, Kind: "arithmetic/add_to_sub", Status: mutation.Runnable}, // duplicate
	}
}

func iterFromSlice(ds []mutation.Descriptor) func(func(mutation.Descriptor) bool) {
	return func(yield func(mutation.Descriptor) bool) {
		for _, d := range ds {
			if !yield(d) {
				return
			}
		}
	}
}

func TestFileExclusionFilter(t *testing.T) {
	f := &FileExclusionFilter{
		Patterns: []*regexp.Regexp{regexp.MustCompile(`_test\.go$`)},
	}

	result := slices.Collect(f.Apply(iterFromSlice(makeDescriptors())))

	for _, d := range result {
		if d.File == "/src/main_test.go" && d.Status != mutation.Skipped {
			t.Errorf("expected test file to be skipped, got %v", d.Status)
		}
		if d.File == "/src/main.go" && d.Status != mutation.Runnable {
			t.Errorf("expected main file to be runnable, got %v", d.Status)
		}
	}
}

func TestFunctionMatchFilter(t *testing.T) {
	f := &FunctionMatchFilter{
		Patterns: []*regexp.Regexp{regexp.MustCompile(`^Foo$`)},
	}

	result := slices.Collect(f.Apply(iterFromSlice(makeDescriptors())))

	for _, d := range result {
		if d.FuncName == "Foo" && d.Status != mutation.Runnable {
			t.Errorf("expected Foo to be runnable, got %v", d.Status)
		}
		if d.FuncName == "Bar" && d.Status != mutation.Skipped {
			t.Errorf("expected Bar to be skipped, got %v", d.Status)
		}
	}
}

func TestEquivalentMutationFilter(t *testing.T) {
	f := &EquivalentMutationFilter{}

	result := slices.Collect(f.Apply(iterFromSlice(makeDescriptors())))

	if len(result) != 4 {
		t.Fatalf("expected 4 results, got %d", len(result))
	}

	equivCount := 0
	for _, d := range result {
		if d.Status == mutation.Equivalent {
			equivCount++
		}
	}
	if equivCount != 1 {
		t.Errorf("expected 1 equivalent, got %d", equivCount)
	}
}

func TestChain(t *testing.T) {
	ds := makeDescriptors()

	result := slices.Collect(Chain(
		iterFromSlice(ds),
		&FileExclusionFilter{Patterns: []*regexp.Regexp{regexp.MustCompile(`_test\.go$`)}},
		&EquivalentMutationFilter{},
	))

	if len(result) != 4 {
		t.Fatalf("expected 4 results after chain, got %d", len(result))
	}

	// The test file should be skipped.
	for _, d := range result {
		if d.File == "/src/main_test.go" && d.Status != mutation.Skipped {
			t.Errorf("expected test file skipped after chain")
		}
	}
}

func TestCoverageFilter(t *testing.T) {
	cov := &CoverageData{
		CoveredLines: map[string]map[int]bool{
			"/src/main.go": {10: true},
		},
	}

	f := &CoverageFilter{Coverage: cov}
	result := slices.Collect(f.Apply(iterFromSlice(makeDescriptors())))

	for _, d := range result {
		if d.File == "/src/main.go" && d.StartPos.Line == 10 && d.Status != mutation.Runnable {
			t.Errorf("covered line should be runnable")
		}
		if d.File == "/src/util.go" && d.Status != mutation.NotCovered {
			t.Errorf("uncovered file should be not_covered, got %v", d.Status)
		}
	}
}

func TestDiffFilter(t *testing.T) {
	diff := &DiffData{
		ChangedLines: map[string]map[int]bool{
			"/src/main.go": {10: true},
		},
	}

	f := &DiffFilter{Diff: diff}
	result := slices.Collect(f.Apply(iterFromSlice(makeDescriptors())))

	for _, d := range result {
		if d.File == "/src/main.go" && d.StartPos.Line == 10 && d.Status != mutation.Runnable {
			t.Errorf("changed line should be runnable")
		}
		if d.File == "/src/util.go" && d.Status != mutation.Skipped {
			t.Errorf("unchanged file should be skipped, got %v", d.Status)
		}
	}
}
