# go-mutesting: Project Status & Modernization Report

> **Date**: February 2026
> **Upstream repository**: [zimmski/go-mutesting](https://github.com/zimmski/go-mutesting)
> **Fork**: [fredbi/go-mutesting](https://github.com/fredbi/go-mutesting)

---

## 1. Current Update Status

### Repository Overview

| Attribute | Value |
|:---|:---|
| **Stars** | 665 |
| **Forks** | 58 |
| **License** | MIT |
| **Last upstream commit** | June 10, 2021 (PR #77 — module-awareness) |
| **Latest release** | v1.2 (June 10, 2021) |
| **Go version in `go.mod`** | 1.10 |
| **CI system** | Travis CI (defunct for OSS since 2021) |
| **Tested Go versions (CI)** | 1.11.x, 1.12.x |

### Release History

| Release | Date | Highlights |
|:---|:---|:---|
| v1.0 | June 25, 2016 | Initial release |
| v1.1 | November 1, 2018 | Comparison mutator, blacklist improvements, Go 1.10/1.11 support, macOS |
| v1.2 | June 10, 2021 | Go module-aware (last release) |

### Activity Assessment

The upstream project has been **effectively dormant since June 2021**. The last commit merged a community PR for Go module support. No maintainer activity has occurred since then — no issue triage, no PR reviews, and no releases in over 4 years.

Issue [#100 — "Is go-mutesting dead? Proposal for fork"](https://github.com/zimmski/go-mutesting/issues/100) (July 2022, 5 comments) explicitly raised the abandonment concern. Notable community forks include [avito-tech/go-mutesting](https://github.com/avito-tech/go-mutesting), though they also appear to have stalled.

---

## 2. Summary of Outstanding Issues and PRs

### Open Issues — 34 total

Issues span from December 2014 to November 2024. They can be categorized as follows:

#### Critical / Bugs (6)
| # | Title | Date |
|:---|:---|:---|
| [#108](https://github.com/zimmski/go-mutesting/issues/108) | Full file diff output instead of specific mutated code location | Nov 2024 |
| [#96](https://github.com/zimmski/go-mutesting/issues/96) | Struct type mutation causes compilation failure | Dec 2021 |
| [#93](https://github.com/zimmski/go-mutesting/issues/93) | Error executing `go-mutesting example/` | Oct 2021 |
| [#82](https://github.com/zimmski/go-mutesting/issues/82) | Stale compilation cache issue | Jan 2021 |
| [#65](https://github.com/zimmski/go-mutesting/issues/65) | Random timeout during mutation testing | Mar 2019 |
| [#54](https://github.com/zimmski/go-mutesting/issues/54) | Panic on single file mutation | Jan 2018 |

#### Feature Requests / Enhancements (20+)
| # | Title | Date | Theme |
|:---|:---|:---|:---|
| [#13](https://github.com/zimmski/go-mutesting/issues/13) | Parallel mutation testing | Jan 2015 | Performance |
| [#37](https://github.com/zimmski/go-mutesting/issues/37) | Coverage-aware mutation testing | Jun 2016 | Performance |
| [#36](https://github.com/zimmski/go-mutesting/issues/36) | Signal handling (SIGINT cleanup) | Jun 2016 | Reliability |
| [#38](https://github.com/zimmski/go-mutesting/issues/38) | Check if test files exist before running | Jun 2016 | UX |
| [#44](https://github.com/zimmski/go-mutesting/issues/44) | Pre-check that tests pass before mutating | Sep 2016 | Reliability |
| [#7](https://github.com/zimmski/go-mutesting/issues/7) | Improve blacklisting (not MD5-based) | Dec 2014 | Correctness |
| [#40](https://github.com/zimmski/go-mutesting/issues/40) | Return value mutation | Jun 2016 | Mutator |
| [#73](https://github.com/zimmski/go-mutesting/issues/73) | Structured output (JSON/HTML reports) | Mar 2020 | Reporting |
| [#67](https://github.com/zimmski/go-mutesting/issues/67) | Include file and line number in output | Mar 2019 | Reporting |
| [#80](https://github.com/zimmski/go-mutesting/issues/80) | Comment-based exclusion (`// nomut`) | Oct 2020 | UX |
| [#85](https://github.com/zimmski/go-mutesting/issues/85) | Skip entire function support | Jun 2021 | UX |
| [#86](https://github.com/zimmski/go-mutesting/issues/86) | CI pipeline integration guidance | Jun 2021 | DevOps |
| [#87](https://github.com/zimmski/go-mutesting/issues/87) | Run only specific mutators | Jun 2021 | UX |
| [#89](https://github.com/zimmski/go-mutesting/issues/89) | Fix score to exclude duplicated/skipped | Aug 2021 | Correctness |
| [#79](https://github.com/zimmski/go-mutesting/issues/79) | Should duplicate mutations count toward score | Oct 2020 | Correctness |

#### Usage Questions / Documentation (8+)
Issues #94, #93, #105, #53, #63, #72, etc. reflect confusion about Go modules support, `GOPATH` vs. module mode, and configuration — indicating documentation and UX gaps.

### Open Pull Requests — 9 total (all unreviewed/stale)

| # | Title | Date | Status |
|:---|:---|:---|:---|
| [#107](https://github.com/zimmski/go-mutesting/pull/107) | Update dependency versions | Jul 2024 | Open, unreviewed |
| [#104](https://github.com/zimmski/go-mutesting/pull/104) | Custom mutation testing | Dec 2022 | Stale |
| [#103](https://github.com/zimmski/go-mutesting/pull/103) | Fix timeout issue (#65) | Nov 2022 | Stale |
| [#102](https://github.com/zimmski/go-mutesting/pull/102) | Add `--only` option for mutators (#87) | Aug 2022 | Stale |
| [#101](https://github.com/zimmski/go-mutesting/pull/101) | Skip whole function (#85) | Jul 2022 | Stale |
| [#99](https://github.com/zimmski/go-mutesting/pull/99) | Show file and line number (#67) | May 2022 | Stale |
| [#98](https://github.com/zimmski/go-mutesting/pull/98) | Add as GitHub Action | Mar 2022 | Stale |
| [#97](https://github.com/zimmski/go-mutesting/pull/97) | Add number literal mutators | Jan 2022 | Stale |
| [#90](https://github.com/zimmski/go-mutesting/pull/90) | Fix score calculation (exclude duplicated) | Sep 2021 | Stale |

A recent closed PR [#110](https://github.com/zimmski/go-mutesting/pull/110) (Apr 2025) attempted to update to Go 1.24 but was closed without merge.

---

## 3. Modernization Plan

### Phase 1 — Foundation (Critical)

| # | Action | Rationale |
|:---|:---|:---|
| 1.1 | **Update `go.mod` to Go 1.22+** (minimum supported, target 1.23/1.24) | Go 1.10 is 7+ years old; modern toolchain features, generics, and standard library improvements require a modern minimum |
| 1.2 | **Replace deprecated APIs**: `io/ioutil` → `os`/`io`, `golang.org/x/lint/golint` → `golangci-lint` | `ioutil` deprecated since Go 1.16; `golint` archived |
| 1.3 | **Update all dependencies** to current versions | `golang.org/x/tools`, `stretchr/testify`, `go-flags`, etc. are 5+ years out of date |
| 1.4 | **Replace Travis CI with GitHub Actions** | Travis CI no longer supports free OSS builds; GitHub Actions is the standard |
| 1.5 | **Adopt `golangci-lint`** as the unified linter, replace `scripts/lint.sh` | Modern Go projects use `golangci-lint` with a `.golangci.yml` configuration |
| 1.6 | **Update installation instructions** from `go get` to `go install` | `go get` for binaries is deprecated since Go 1.17 |

### Phase 2 — Code Quality & Reliability

| # | Action | Rationale |
|:---|:---|:---|
| 2.1 | **Replace `md5` blacklisting** with file-path + line-range based exclusions | Issue #7 (2014); MD5 checksums break on any source change |
| 2.2 | **Add signal handling** (`SIGINT`, `SIGTERM`) with proper cleanup | Issue #36; mutations leave files in corrupted state on interrupt |
| 2.3 | **Pre-flight test validation** — verify tests pass before mutating | Issue #44; failing tests make all mutations appear killed |
| 2.4 | **Replace `fmt.Errorf` patterns** with `%w` error wrapping | Modern Go error handling best practice (Go 1.13+) |
| 2.5 | **Add `context.Context`** to exec commands for proper timeout/cancellation | Replace ad-hoc timeout with `exec.CommandContext` |

### Phase 3 — Performance & Features

| # | Action | Rationale |
|:---|:---|:---|
| 3.1 | **Implement parallel mutation testing** with worker pool | Issue #13 (2015); sequential testing is prohibitively slow for real codebases |
| 3.2 | **Add coverage-aware mutation** — skip mutating uncovered lines | Issue #37; reduces wasted work significantly |
| 3.3 | **Add structured output** — JSON, JUnit XML for CI integration | Issue #73; enables integration with CI dashboards |
| 3.4 | **Add file/line number** to mutation output | Issue #67, PR #99; essential for developer workflow |
| 3.5 | **Add new mutators** — return values, number literals, string mutations | Issues #40, PR #97; increases mutation coverage |
| 3.6 | **Publish as GitHub Action** | PR #98; simplifies CI adoption |

### Phase 4 — Community & Documentation

| # | Action | Rationale |
|:---|:---|:---|
| 4.1 | **Refresh README** with modern examples, `go install` instructions | Current README references `GOPATH` workflow |
| 4.2 | **Add CONTRIBUTING.md** and issue templates | Facilitate community contributions |
| 4.3 | **Triage and close stale issues/PRs** on the fork | 34 open issues, 9 stale PRs need resolution |
| 4.4 | **Add GoDoc badges and pkg.go.dev links** | Replace defunct godoc.org links |

---

## 4. Top 3 Major Shortcomings and Remediation

### Shortcoming 1: Severely Outdated Go Version and Dependencies

**Impact**: The project declares `go 1.10` in `go.mod`, uses APIs deprecated since Go 1.16 (`io/ioutil`), depends on archived tools (`golang.org/x/lint`), and relies on defunct CI (Travis CI). It cannot be built with `go install` using modern Go toolchains without warnings. Dependencies have not been updated since 2019, potentially carrying unpatched vulnerabilities.

**Evidence**:
- `go.mod`: `go 1.10`
- `parse.go`: uses `ioutil.ReadFile`
- `cmd/go-mutesting/main.go`: uses `ioutil.TempDir`, `ioutil.ReadFile`, `ioutil.WriteFile`
- `.travis.yml`: Travis CI with Go 1.11/1.12
- `Makefile`: uses `go get` for tool installation
- `scripts/lint.sh`: uses `golint` (archived)

**Remediation**:
1. Update `go.mod` to `go 1.22` minimum (supporting the two most recent Go releases per Go policy).
2. Run `go mod tidy` to clean dependency graph; update all `golang.org/x/*` and third-party dependencies.
3. Replace `ioutil` calls with their `os`/`io` equivalents.
4. Replace Travis CI with a GitHub Actions workflow (`.github/workflows/ci.yml`).
5. Replace `golint`+`errcheck`+`staticcheck` scripts with `golangci-lint`.
6. Update README installation instructions to use `go install`.

### Shortcoming 2: No Parallel Mutation Testing (Performance)

**Impact**: Mutation testing is inherently CPU-intensive — every mutation requires recompilation and test execution. The current implementation processes mutations **strictly sequentially**, making it impractical for real-world codebases. This is the oldest open feature request (issue #13, January 2015) and the single biggest barrier to adoption.

**Evidence**:
- `walk.go`: `MutateWalk()` uses a single goroutine with a blocking channel protocol.
- `cmd/go-mutesting/main.go`: `mutate()` processes each mutation in a single loop.
- `mutateExec()`: replaces the original file in-place, preventing concurrent execution.

**Remediation**:
1. Decouple mutation generation from mutation testing — generate all mutations first, then test in parallel.
2. Implement a worker pool pattern (using `sync.WaitGroup` or `golang.org/x/sync/errgroup`).
3. Isolate each mutation test in a temporary directory (copy the package under test) so workers don't conflict.
4. Add a `--workers` / `-j` flag to control concurrency (default: `runtime.NumCPU()`).
5. Implement proper `context.Context` propagation for timeout and cancellation across workers.

### Shortcoming 3: Poor Mutation Reporting and Developer Experience

**Impact**: The tool outputs free-text diffs to stdout with no structured format, no file/line numbers, and no integration hooks for CI systems. This makes it difficult to act on results, integrate with code review workflows, or track mutation scores over time. Multiple issues (#67, #73, #80, #86, #89) and PRs (#99) highlight this gap.

**Evidence**:
- `main.go` line 303: `fmt.Printf("The mutation score is %f (%d passed, ...")` — unstructured text.
- Diff output uses external `diff -u` with no file/line metadata.
- Score calculation includes skipped/duplicated mutations, which inflates failure appearance (issue #89).
- No support for JSON, JUnit XML, or HTML output.
- No comment-based exclusion mechanism (`// nomut` or similar).

**Remediation**:
1. Add an `--output-format` flag supporting `text` (default), `json`, and `junit-xml`.
2. Include file path, line number, function name, and mutator name in each mutation result.
3. Fix score calculation to exclude duplicated mutations from total (matching PR #90).
4. Add `// nomut` comment support to skip specific lines or functions.
5. Produce a summary report suitable for CI consumption (exit code based on score threshold via `--score-threshold`).

---

## 5. Notable Community Forks

| Fork | Last Activity | Notes |
|:---|:---|:---|
| [avito-tech/go-mutesting](https://github.com/avito-tech/go-mutesting) | ~2022 | Added some features, also appears stalled |
| [buzyka/go-mutesting](https://github.com/buzyka/go-mutesting) | Apr 2025 | Attempted Go 1.24 update (PR #110, closed) |
| [andrealim11/go-mutesting](https://github.com/andrealim11/go-mutesting) | Jul 2024 | Dependency update (PR #107, open) |

---

## 6. Conclusion

`go-mutesting` is a well-designed mutation testing framework with a solid AST-based architecture, but it has been unmaintained for over 4 years. The project requires significant modernization to be viable with current Go (1.22+). The three most impactful improvements are:

1. **Updating to modern Go** (dependencies, APIs, CI, tooling)
2. **Adding parallel mutation testing** (performance)
3. **Improving reporting and developer experience** (adoption)

The recommended approach for this fork is to tackle Phase 1 (foundation) immediately, as it is prerequisite for all other improvements and restores basic functionality with current Go toolchains.
