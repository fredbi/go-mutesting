package walker_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/fredbi/go-mutesting/pkg/applier"
	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/pruning"
	"github.com/fredbi/go-mutesting/pkg/walker"

	_ "github.com/fredbi/go-mutesting/internal/strategies/all"
)

func TestIntegrationDiscoverPruneApply(t *testing.T) {
	const testFile = "../../testdata/sample.go"

	fset, file, pkg, info, src := parseAndTypeCheck(t, testFile)

	// Step 1: Discover all mutations.
	w := walker.New(nil)
	all := slices.Collect(w.Discover(fset, file, pkg, info, src, testFile, "sample"))
	t.Logf("discovered %d mutations", len(all))

	if len(all) == 0 {
		t.Fatal("no mutations discovered")
	}

	// Step 2: Prune (dedup only for this test).
	pruned := slices.Collect(pruning.Chain(
		slices.Values(all),
		&pruning.EquivalentMutationFilter{},
	))

	runnable := 0
	for _, d := range pruned {
		if d.Status == mutation.Runnable {
			runnable++
		}
	}
	t.Logf("runnable after pruning: %d", runnable)

	// Step 3: Apply a token mutation, verify file compiles.
	var tokenDesc *mutation.Descriptor
	for i, d := range pruned {
		if d.Apply.TokenSwap != nil && d.Status == mutation.Runnable {
			tokenDesc = &pruned[i]
			break
		}
	}

	if tokenDesc == nil {
		t.Fatal("no runnable token mutation found")
	}

	workdir := t.TempDir()
	destFile := filepath.Join(workdir, testFile)
	if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destFile, src, 0o644); err != nil {
		t.Fatal(err)
	}

	composite := applier.NewCompositeApplier()

	if err := composite.Apply(*tokenDesc, workdir); err != nil {
		t.Fatalf("Apply token mutation: %v", err)
	}

	mutatedSrc, _ := os.ReadFile(destFile)
	t.Logf("applied mutation %s (%s -> %s)", tokenDesc.Kind, tokenDesc.Original, tokenDesc.Replacement)

	if string(mutatedSrc) == string(src) {
		t.Error("mutated file should differ from original")
	}

	// Verify the mutated token is present.
	if !strings.Contains(string(mutatedSrc), tokenDesc.Replacement) {
		t.Errorf("mutated file should contain replacement %q", tokenDesc.Replacement)
	}

	// Rollback and verify.
	if err := composite.Rollback(*tokenDesc, workdir); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	restoredSrc, _ := os.ReadFile(destFile)
	if !strings.Contains(string(restoredSrc), tokenDesc.Original) {
		t.Errorf("restored file should contain original %q", tokenDesc.Original)
	}
}
