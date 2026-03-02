# Comparative Analysis: go-mutesting vs. gremlins

> **Date**: March 2026
> **Repositories compared**:
> - [fredbi/go-mutesting](https://github.com/fredbi/go-mutesting) (fork of [zimmski/go-mutesting](https://github.com/zimmski/go-mutesting))
> - [go-gremlins/gremlins](https://github.com/go-gremlins/gremlins)

---

## 1. Supported Mutators — Comparative Table

| Mutation Category | go-mutesting | gremlins | Notes |
|:---|:---|:---|:---|
| **Arithmetic operators** (`+` ↔ `-`, `*` ↔ `/`, `%` → `*`) | ❌ | ✅ `ArithmeticBase` | gremlins mutates all basic arithmetic operators |
| **Comparison boundary** (`<` ↔ `<=`, `>` ↔ `>=`) | ✅ `expression/comparison` | ✅ `ConditionalsBoundary` | Both catch off-by-one errors; gremlins also covers `==`/`!=` boundaries |
| **Comparison negation** (`==` ↔ `!=`, `<` ↔ `>=`, `>` ↔ `<=`) | ❌ | ✅ `ConditionalsNegation` | gremlins fully inverts conditionals; go-mutesting only shifts boundaries |
| **Increment/Decrement** (`++` ↔ `--`) | ❌ | ✅ `IncrementDecrement` | — |
| **Assignment inversion** (`+=` ↔ `-=`, `*=` ↔ `/=`) | ❌ | ✅ `InvertAssignments` | — |
| **Bitwise operator inversion** (`&` ↔ `\|`, `^` → `&`, `<<` ↔ `>>`) | ❌ | ✅ `InvertBitwise` | — |
| **Bitwise assignment inversion** (`&=` ↔ `\|=`, `<<=` ↔ `>>=`) | ❌ | ✅ `InvertBitwiseAssignments` | — |
| **Logical operator inversion** (`&&` ↔ `\|\|`) | ❌ | ✅ `InvertLogical` | go-mutesting's `expression/remove` is related but distinct (see below) |
| **Loop control inversion** (`break` ↔ `continue`) | ❌ | ✅ `InvertLoopCtrl` | — |
| **Negate numeric values** (unary `-` → `+`) | ❌ | ✅ `InvertNegatives` | — |
| **Remove self-assignments** (`+=` → `=`, `\|=` → `=`, etc.) | ❌ | ✅ `RemoveSelfAssignments` | Replaces compound assignments with plain assignment |
| **Remove logical terms** (`a && b` → `true`/`b`, `a \|\| b` → `false`/`a`) | ✅ `expression/remove` | ❌ | go-mutesting replaces each operand with `true` or `false` |
| **Remove statements** (assignments, expressions, inc/dec) | ✅ `statement/remove` | ❌ | go-mutesting replaces statements with a no-op (`_ = 0`) |
| **Empty if body** | ✅ `branch/if` | ❌ | go-mutesting replaces if/else-if body with no-op |
| **Empty else body** | ✅ `branch/else` | ❌ | go-mutesting replaces else body with no-op |
| **Empty case body** | ✅ `branch/case` | ❌ | go-mutesting replaces switch case body with no-op |

### Summary Count

| Metric | go-mutesting | gremlins |
|:---|:---|:---|
| **Total mutator types** | 6 | 11 |
| **Operator-level mutators** | 2 | 9 |
| **Statement/block-level mutators** | 4 | 2 (via operator removal) |

---

## 2. Qualitative Assessment of Mutator Implementations

### go-mutesting: AST-node mutators

go-mutesting operates at the **AST node level**. Each mutator implements a `Mutator` interface with `MutateWalk`, which
traverses the full AST using `ast.Walk` and yields mutations via a channel-based protocol (`Change` / `Reset`).

**Strengths**:
- **Structural mutations**: Can perform complex structural changes such as emptying entire if/else/case blocks and removing
  complete statements — mutations that are impossible at the token level.
- **Type-aware**: Uses `go/types` for type checking (e.g., the statement remover validates that removing a statement
  won't cause a compilation error by checking for `:=` declarations).
- **Pluggable architecture**: New mutators are registered via `mutator.Register()` and discovered automatically.

**Weaknesses**:
- **Limited operator mutations**: Only covers comparison boundary shifts (`<` ↔ `<=`, `>` ↔ `>=`). Does not mutate
  arithmetic, logical, bitwise, or assignment operators.
- **No conditional negation**: Cannot invert equality or inequality operators.
- **No increment/decrement or loop control mutations**.

### gremlins: Token-level mutators

gremlins operates at the **token level**. It walks the AST to find token positions, then looks up a static mapping
(`TokenMutantType`) from each `token.Token` to applicable `mutator.Type` values. The actual mutation is a token
replacement defined in the `tokenMutations` map.

**Strengths**:
- **Broad operator coverage**: 11 mutator types covering arithmetic, comparison, logical, bitwise, assignment,
  increment/decrement, loop control, and negation operators.
- **Simple and systematic**: The token-swap approach is deterministic and exhaustive for operator-level mutations.
- **Concurrency-safe**: Uses per-file locks when applying mutations to the shared AST, enabling safe parallel execution.

**Weaknesses**:
- **No structural mutations**: Cannot empty blocks (if/else/case bodies) or remove statements. Mutations are limited to
  single-token replacements.
- **No expression-level removal**: Cannot test whether entire sub-expressions are necessary (e.g., removing an operand
  of `&&` / `||`).
- **Less type-aware**: Does not use `go/types` for semantic validation; relies on compilation failure detection
  (`NotViable` status) as a post-hoc filter.

### Complementarity

The two tools have **largely complementary** mutation strategies:

| Dimension | go-mutesting | gremlins |
|:---|:---|:---|
| **Mutation granularity** | AST node (block/statement) | Token (operator) |
| **Operator coverage** | Narrow (comparisons only) | Broad (11 operator categories) |
| **Structural coverage** | Strong (if/else/case/statement removal) | None |
| **Type awareness** | Yes (`go/types`) | Minimal (compilation check) |

A hypothetical ideal mutation testing tool would combine both approaches: gremlins' exhaustive operator mutations with
go-mutesting's structural block/statement mutations.

---

## 3. How Mutations Are Built and Run

### go-mutesting

**Mutation approach**: In-place AST modification.

1. **Parse**: Source file is parsed into AST with `go/parser` and type-checked with `go/types`.
2. **Walk**: `MutateWalk()` traverses the AST. For each mutable node, the mutator calls `Change()` (modifying the AST in-place) and signals readiness via a channel.
3. **Generate**: The modified AST is printed back to a temporary file using `go/printer`.
4. **Test**: The original source file is replaced with the mutated version, and `go test` is run on the package.
5. **Restore**: `Reset()` reverts the AST to its original state; the original file is restored.

**Execution model**: Built-in executor runs `go test` directly. Alternatively, a user-provided exec command receives
environment variables (`MUTATE_CHANGED`, `MUTATE_ORIGINAL`, `MUTATE_PACKAGE`, `MUTATE_TIMEOUT`, etc.) and returns an
exit code (0 = killed, 1 = lived, 2 = skipped).

**Key characteristic**: The original file is **replaced in-place**, making concurrent mutation testing unsafe without
directory isolation.

### gremlins

**Mutation approach**: Token substitution in a shared AST, applied to isolated working directory copies.

1. **Discover**: Engine walks the AST of all non-test `.go` files. For each token, checks `TokenMutantType` to find applicable mutation types.
2. **Classify**: Each candidate mutation is classified as `Runnable` (covered by tests), `NotCovered` (not in coverage profile), or `Skipped` (outside diff scope).
3. **Queue**: Mutations are streamed into a worker pool via a channel.
4. **Apply**: Each worker operates on its own copy of the source tree (via `workdir.Dealer`). The token is swapped in the shared AST, the modified file is written to the worker's directory, and the token is immediately restored.
5. **Test**: `go test` is run with a context-based timeout (default: 5× coverage execution time).
6. **Rollback**: The original file content (saved before mutation) is restored in the worker's directory.

**Key characteristic**: Each worker gets its own working directory copy, enabling **safe parallel execution**.

| Aspect | go-mutesting | gremlins |
|:---|:---|:---|
| **Mutation level** | AST node | Token |
| **File handling** | In-place replacement of original | Isolated working directory per worker |
| **Test execution** | `go test` or custom exec command | `go test` only |
| **Timeout management** | `--timeout` flag passed to exec | `context.WithTimeout` (5× coverage time) |
| **Custom execution** | ✅ via `--exec` with env vars | ❌ `go test` only |
| **Build verification** | Tests compilation first, then runs tests | Relies on test failure exit codes |

---

## 4. Mutation Pruning

### go-mutesting

| Pruning Strategy | Mechanism | Effectiveness |
|:---|:---|:---|
| **Blacklist (MD5)** | `--blacklist` file containing MD5 checksums of known false-positive mutations | Low — checksums change with any source code modification (acknowledged as "badly implemented" in README) |
| **Function filter** | `--match` regex to restrict mutations to matching function names | Medium — useful for targeting specific areas |
| **Mutator disable** | `--disable` pattern to turn off specific mutator types (supports wildcards) | Medium — reduces noise from unwanted mutation categories |
| **Statement validation** | `checkRemoveStatement()` verifies that a statement can be safely removed (skips `:=` declarations, control flow) | High — prevents compile-time failures at mutation generation time |
| **Coverage-based pruning** | ❌ Not implemented | — (Issue #37 requested this) |
| **Diff-based pruning** | ❌ Not implemented | — |
| **File exclusion patterns** | ❌ Not implemented | — |

### gremlins

| Pruning Strategy | Mechanism | Effectiveness |
|:---|:---|:---|
| **Coverage-based pruning** | `coverage.Profile.IsCovered(pos)` checks if the mutation position is covered by tests; uncovered mutations are marked `NotCovered` and skipped | High — avoids testing mutations that no test can detect, reducing wasted test runs significantly |
| **Diff-based pruning** | `diff.Diff.IsChanged(pos)` checks if the mutation position is within changed files; unchanged positions are marked `Skipped` | High — in CI pipelines, only tests mutations relevant to the current changeset |
| **File exclusion** | Regex patterns via `exclude-files` configuration to skip files | Medium — useful for generated code, vendored dependencies |
| **Mutator type toggle** | Each mutator type can be individually enabled/disabled via configuration | Medium — same as go-mutesting's `--disable` |
| **Compilation failure detection** | `NotViable` status for mutations that cause build failures (exit code 2) | Reactive — detected after attempting to test, not before |
| **Blacklist/MD5** | ❌ Not implemented | — |
| **Function filter** | ❌ Not implemented | — |

### Pruning Comparison

| Strategy | go-mutesting | gremlins |
|:---|:---|:---|
| Coverage-aware skipping | ❌ | ✅ |
| Diff-based filtering | ❌ | ✅ |
| File exclusion patterns | ❌ | ✅ |
| Function name filter | ✅ | ❌ |
| MD5 blacklist | ✅ (fragile) | ❌ |
| Type-aware validation | ✅ | ❌ (post-hoc) |
| Mutator toggle | ✅ | ✅ |

gremlins' pruning is significantly more sophisticated. Coverage-based and diff-based pruning can reduce the number of
mutations to test by an order of magnitude in real-world CI workflows, which is critical given that mutation testing is
inherently expensive.

---

## 5. Parallelism

### go-mutesting: Sequential only

go-mutesting processes mutations **strictly sequentially**:

- Files are iterated one by one.
- For each file, mutators are applied one by one.
- For each mutation, `Change()` → test → `Reset()` happens in a blocking loop.
- The original source file is replaced in-place, making concurrent mutations on the same package unsafe.

The channel protocol in `MutateWalk()` coordinates the single AST walker goroutine with the main execution loop — this
is a **synchronization mechanism**, not parallelism.

**Parallelism support**: ❌ None. Issue #13 (January 2015) is the oldest open feature request.

### gremlins: Worker pool parallelism

gremlins implements **full parallel mutation testing** via a worker pool:

- **Worker pool**: `workerpool.Pool` manages N concurrent workers (default: `runtime.NumCPU()`).
- **Configurable**: The number of workers can be set via `UnleashWorkersKey` configuration.
- **Integration mode**: In integration mode, worker count is halved (`wNum / 2`) to reduce resource contention.
- **Isolation**: Each worker operates on its own working directory copy (managed by `workdir.Dealer`), preventing file conflicts.
- **Per-file locking**: The `TokenMutator` uses a per-file mutex to prevent concurrent AST modifications to the same file, while allowing parallel mutations on different files.
- **Context-based cancellation**: The engine propagates `context.Context` for graceful timeout and cancellation.

**Parallelism support**: ✅ Full. Default `runtime.NumCPU()` workers with per-worker directory isolation.

### Parallelism Comparison

| Aspect | go-mutesting | gremlins |
|:---|:---|:---|
| **Parallel execution** | ❌ Sequential | ✅ Worker pool |
| **Default concurrency** | 1 | `runtime.NumCPU()` |
| **Configurable workers** | N/A | ✅ via configuration |
| **File isolation** | In-place (unsafe for concurrency) | Per-worker directory copy |
| **AST safety** | N/A (single-threaded) | Per-file mutex locks |
| **Cancellation** | None (no context support) | `context.Context` with timeout |
| **Integration mode** | N/A | Workers ÷ 2 |

---

## 6. Overall Assessment

| Dimension | go-mutesting | gremlins |
|:---|:---|:---|
| **Maturity** | Older (2014), dormant since 2021 | Newer (2022), actively maintained |
| **Mutator breadth** | 6 types (structural focus) | 11 types (operator focus) |
| **Unique strengths** | Block emptying, statement removal, custom exec commands, function-level filtering | Coverage/diff pruning, parallelism, broad operator mutations |
| **Performance** | Slow (sequential, no pruning) | Fast (parallel, coverage-pruned) |
| **Extensibility** | Pluggable mutator interface | Configuration-driven toggle |
| **CI integration** | Basic (exit code + text output) | Better (structured reporting, diff-aware) |
| **Custom test commands** | ✅ `--exec` with env vars | ❌ `go test` only |

### Key Takeaways

1. **gremlins has broader operator mutation coverage** (11 vs. 6 types), but **go-mutesting has unique structural mutations** (block emptying, statement removal) that gremlins cannot perform.

2. **gremlins is significantly faster** due to parallel execution and coverage-based pruning. go-mutesting's sequential, unfiltered approach makes it impractical for large codebases.

3. **The tools are complementary rather than competitive**: an ideal mutation testing tool for Go would combine gremlins' operator mutations and pruning with go-mutesting's structural mutations and custom exec support.

4. **go-mutesting's `--exec` interface** is a unique advantage — it allows integration with arbitrary test frameworks and build systems, whereas gremlins is tightly coupled to `go test`.

5. **For CI adoption**, gremlins is more practical today due to parallelism, diff-based pruning, and structured output. go-mutesting would need the improvements outlined in the [modernization report](STATUS_AND_MODERNIZATION_REPORT.md) to compete.
