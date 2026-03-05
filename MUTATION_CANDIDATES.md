# Mutation Candidates (Deferred)

This document lists mutation ideas that were considered but **not implemented** in the current pipeline.
They are recorded here for future reference and potential implementation in later iterations.

Sources: stryker4s (Scala mutation testing), general mutation testing literature.

---

## From stryker4s Analysis

### Richer Equality Cross-Mutations

**Priority: Low**

stryker4s mutates comparison operators more aggressively than we do, e.g. `>` to `<`, `>` to `==`, `>=` to `!=`.
Our `conditional/boundary` and `conditional/negation` strategies already cover the most important inversions:

| We have | stryker4s also does |
|:--------|:--------------------|
| `<` <-> `<=` (boundary) | `<` -> `>` (cross) |
| `<` -> `>=` (negation) | `<` -> `==` (cross) |
| `>` <-> `>=` (boundary) | `>` -> `<` (cross) |
| `>` -> `<=` (negation) | `>` -> `==` (cross) |
| `==` <-> `!=` (negation) | `>=` -> `!=` (cross) |

**Judgement:** Our boundary + negation coverage is sufficient for most real-world Go code. The cross-mutations
(`>` -> `<`, `>` -> `==`) are more likely to produce equivalent mutants or trivially killed mutants.
Adding them would increase mutant count substantially without proportional improvement in test quality assessment.

Could be reconsidered if users report false confidence (tests pass but meaningful comparison bugs are missed).

---

### Method Expression Swaps

**Priority: Not Applicable**

stryker4s replaces Scala collection methods: `filter` <-> `filterNot`, `exists` <-> `forall`,
`take` <-> `drop`, `min` <-> `max`, `indexOf` <-> `lastIndexOf`, `isEmpty` <-> `nonEmpty`.

**Judgement:** Not applicable to Go. Go does not have these as built-in methods on standard types.
The closest Go equivalents would be swapping `slices.Min` <-> `slices.Max` or `strings.Contains` <-> `!strings.Contains`,
but these are free functions, not method calls, and the patterns are too diverse to enumerate meaningfully.

If a future version supports user-defined swap rules (e.g., via config), this category could be revisited
as user-supplied pairs.

---

### Regex Pattern Mutations

**Priority: Low**

stryker4s mutates regex patterns: `^` -> empty, `$` -> empty, `*` -> `?` (greedy to lazy),
`+` -> `*`, `\d` -> `\D`, etc.

**Judgement:** Complex to implement correctly in Go (requires regex-aware parsing within string literals).
Niche use case: most Go projects have relatively few regex patterns, and those that do often test them
via integration or golden-file tests. High risk of producing non-compiling mutants (invalid regex at runtime)
which would need to be filtered post-execution.

Could be implemented as an optional strategy behind a flag if there's user demand.

---

## Go-Specific Ideas (Not Yet Implemented)

### Goroutine Removal

**Priority: Medium**

Remove `go` keyword from `go func()` calls, turning concurrent calls into synchronous ones.
Would test whether tests actually verify concurrent behavior.

**Concerns:** May cause deadlocks if the goroutine is expected to run concurrently (e.g., channel consumers).
Would need careful filtering to avoid test hangs.

### Defer Removal

**Priority: Medium**

Remove `defer` keyword, causing cleanup to execute immediately instead of at function exit.
Would test whether tests verify cleanup ordering.

**Concerns:** Similar to goroutine removal -- may cause panics or resource issues that aren't meaningful
mutation testing signals. Best combined with a timeout mechanism.

### Channel Direction Mutation

**Priority: Low**

Mutate channel operations: send <-> receive, buffered size changes, close removal.
Very Go-specific but niche -- most code doesn't use channels extensively.

### Interface Nil Check Removal

**Priority: Medium**

Remove nil checks on interface values (`if err != nil` is covered by `return/nil_error`,
but other interface nil checks like `if handler != nil { handler.Handle() }` are not).
Could be a specialization of the `branch/empty_if` strategy that specifically targets nil guards.

### Error Wrapping Mutation

**Priority: Low**

Replace `fmt.Errorf("...: %w", err)` with `err` (remove wrapping), or replace `%w` with `%v`
(remove unwrap support). Very Go-idiomatic but niche -- tests for error wrapping are uncommon
outside of library code.

---

## Implementation Notes

When implementing any of these candidates:

1. Follow the existing strategy pattern: implement `strategy.Strategy`, self-register via `init()`
2. Add the strategy package to `internal/strategies/all/all.go`
3. Define new `mutation.Kind` constants in `pkg/mutation/kind.go`
4. Write unit tests following the existing pattern (parse source, discover, assert descriptors)
5. Update `MUTATIONS.md` with the new mutations
6. Consider equivalent mutant risk -- mutations that are frequently equivalent waste CI time
