# Creating a New Mutation Strategy

A step-by-step guide for adding a new mutation strategy to go-mutesting v2.

## When to Use This Skill

Use this skill when you need to:
- Add a new mutation operator (token-swap or structural)
- Port a mutation from gremlins, stryker4s, or another mutation testing tool
- Extend the mutation catalogue with a new class of mutations

## Checklist

Every new mutation strategy requires these steps in order:

1. **Add `Kind` constant(s)** in `pkg/mutation/kind.go`
2. **Add `StructuralAction`** in `pkg/mutation/descriptor.go` (structural mutations only, if no existing action fits)
3. **Create strategy package** in `internal/strategies/<name>/`
4. **Write tests** in `internal/strategies/<name>/<name>_test.go`
5. **Add blank import** in `internal/strategies/all/all.go`
6. **Add applier handler** in `pkg/applier/structural.go` (only if step 2 added a new action)
7. **Update `MUTATIONS.md`** with the new rules
8. **Update `MayHang()`** in `pkg/mutation/kind.go` if the mutation may cause deadlocks

## Two Mutation Patterns

### Pattern A: Token-Swap (byte splice)

For mutations that replace one token with another (operators, keywords, literals).
No AST re-parsing needed at apply time — just byte-splice the file.

**Use when:** the mutation is a textual substitution at a known byte offset.

**Example:** `+` to `-`, `&&` to `||`, `break` to `continue`, `true` to `false`.

**Template** (see `internal/strategies/arithmetic/arithmetic.go`):

```go
package mystrategy

import (
    "go/ast"
    "go/token"
    "iter"

    "github.com/fredbi/go-mutesting/pkg/mutation"
    "github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
    strategy.Register(&myStrategy{})
}

var mutations = []struct {
    from token.Token
    to   token.Token
    kind mutation.Kind
}{
    {token.ADD, token.SUB, mutation.MyKindAddToSub},
    // ...
}

type myStrategy struct{}

func (s *myStrategy) Name() string       { return "mystrategy" }
func (s *myStrategy) NodeTypes() []string { return []string{"*ast.BinaryExpr"} }

func (s *myStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
    return func(yield func(mutation.Descriptor) bool) {
        n, ok := node.(*ast.BinaryExpr)
        if !ok {
            return
        }

        for _, m := range mutations {
            if n.Op != m.from {
                continue
            }

            pos := ctx.Fset.Position(n.OpPos)
            endOffset := pos.Offset + len(n.Op.String())

            desc := mutation.Descriptor{
                File:    ctx.FilePath,
                PkgPath: ctx.PkgPath,
                StartPos: mutation.Position{
                    Line: pos.Line, Column: pos.Column, Offset: pos.Offset,
                },
                EndPos: mutation.Position{
                    Line: pos.Line, Column: pos.Column + len(n.Op.String()), Offset: endOffset,
                },
                Kind:        m.kind,
                Status:      mutation.Runnable,
                Original:    m.from.String(),
                Replacement: m.to.String(),
                Apply: mutation.ApplySpec{
                    TokenSwap: &mutation.TokenSwapSpec{
                        OriginalToken:    m.from.String(),
                        ReplacementToken: m.to.String(),
                        StartOffset:      pos.Offset,
                        EndOffset:        endOffset,
                    },
                },
            }

            if !yield(desc) {
                return
            }
        }
    }
}
```

### Pattern B: Structural (AST re-parse + modify)

For mutations that modify the AST structure: removing statements, emptying blocks, swapping branches, offsetting expressions.

**Use when:** the mutation cannot be expressed as a simple text replacement, or when the applier needs to generate type-aware noop statements.

**Example:** remove a statement, empty an if-body, swap if/else, offset `s[i]` to `s[i+1]`.

**Template** (see `internal/strategies/chanops/chanops.go`):

```go
package mystrategy

import (
    "go/ast"
    "iter"

    "github.com/fredbi/go-mutesting/pkg/mutation"
    "github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
    strategy.Register(&myStrategy{})
}

type myStrategy struct{}

func (s *myStrategy) Name() string       { return "mystrategy" }
func (s *myStrategy) NodeTypes() []string { return []string{"*ast.BlockStmt", "*ast.CaseClause"} }

func (s *myStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
    return func(yield func(mutation.Descriptor) bool) {
        var stmts []ast.Stmt
        switch n := node.(type) {
        case *ast.BlockStmt:
            stmts = n.List
        case *ast.CaseClause:
            stmts = n.Body
        default:
            return
        }

        for i, stmt := range stmts {
            if !matchesMyCondition(stmt) {
                continue
            }

            stmtStart := ctx.Fset.Position(stmt.Pos())
            stmtEnd := ctx.Fset.Position(stmt.End())

            desc := mutation.Descriptor{
                File:    ctx.FilePath,
                PkgPath: ctx.PkgPath,
                StartPos: mutation.Position{
                    Line: stmtStart.Line, Column: stmtStart.Column, Offset: stmtStart.Offset,
                },
                EndPos: mutation.Position{
                    Line: stmtEnd.Line, Column: stmtEnd.Column, Offset: stmtEnd.Offset,
                },
                Kind:        mutation.MyStructuralKind,
                Status:      mutation.Runnable,
                Original:    sourceText(ctx.Src, stmtStart.Offset, stmtEnd.Offset),
                Replacement: "noop (remove my thing)",
                Apply: mutation.ApplySpec{
                    Structural: &mutation.StructuralSpec{
                        NodeType:    "BlockStmt",
                        Action:      mutation.ActionRemoveStatement,
                        TargetIndex: i,
                        StartOffset: stmtStart.Offset,
                        EndOffset:   stmtEnd.Offset,
                    },
                },
            }

            if !yield(desc) {
                return
            }
        }
    }
}
```

## Choosing the Right AST Node Type

The `NodeTypes()` return value determines which AST nodes the walker dispatches to your strategy.

| You want to mutate... | Dispatch on | Why |
|:---|:---|:---|
| Binary operators (`+`, `<`, `&&`) | `*ast.BinaryExpr` | Direct access to `Op`, `OpPos` |
| Unary operators (`-x`, `!x`) | `*ast.UnaryExpr` | Direct access to `Op` |
| Assignments (`+=`, `-=`) | `*ast.AssignStmt` | Direct access to `Tok` |
| `i++` / `i--` | `*ast.IncDecStmt` | Direct access to `Tok` |
| `break` / `continue` | `*ast.BranchStmt` | Direct access to `Tok` |
| If/else bodies | `*ast.IfStmt` | Access to `Body`, `Else` |
| Case bodies | `*ast.CaseClause` | Access to `Body` |
| Statements in a block | `*ast.BlockStmt`, `*ast.CaseClause` | Iterate `List`/`Body` with index for `ActionRemoveStatement` |
| Function calls | `*ast.CallExpr` | Access to `Fun`, `Args` |
| Index expressions | `*ast.IndexExpr` | Access to `Index` |
| Slice expressions | `*ast.SliceExpr` | Access to `Low`, `High` |
| Return statements | `*ast.ReturnStmt` | Access to `Results` |
| Boolean/string literals | `*ast.Ident`, `*ast.BasicLit` | Direct value access |

**Key rule:** If you need the statement's **index within its parent block** (for `ActionRemoveStatement`), dispatch on `*ast.BlockStmt`/`*ast.CaseClause` and iterate statements yourself. Otherwise, dispatch directly on the node you care about.

## Reusable StructuralActions

Before adding a new `StructuralAction`, check if an existing one fits:

| Action | What it does | Used by |
|:---|:---|:---|
| `ActionEmptyBlock` | Replace block body with noop | branch (empty_if, empty_else, empty_case) |
| `ActionRemoveStatement` | Replace statement with type-aware noop (`_ = x`) | statement, lockremove, chanops |
| `ActionReplaceWithTrue` | Replace expression with `true` | condexpr, nilguard |
| `ActionReplaceWithFalse` | Replace expression with `false` | condexpr, nilguard |
| `ActionSwapIfElse` | Swap if-body and else-body | swapbranch |
| `ActionSwapCase` | Swap two adjacent case clause bodies | swapbranch |
| `ActionReturnZero` | Replace return values with zero values | returnval |
| `ActionNegateBoolReturn` | Negate boolean return value | returnval |
| `ActionReplaceStmtWithReturn` | Replace a statement with a return | panicremove |
| `ActionSwapCallArgs` | Swap two arguments in a function call | argswap |
| `ActionOffsetExpr` | Wrap expression in `expr + 1` / `expr - 1` | slicebound |

## Test Pattern

All strategy tests follow the same pattern: parse source, type check, iterate registered strategies, collect descriptors.

```go
package mystrategy_test

import (
    "go/ast"
    "go/parser"
    "go/token"
    "go/types"
    "slices"
    "testing"

    "github.com/fredbi/go-mutesting/pkg/mutation"
    "github.com/fredbi/go-mutesting/pkg/strategy"

    _ "github.com/fredbi/go-mutesting/internal/strategies/mystrategy"
)

func discoverNodes(t *testing.T, src string) []mutation.Descriptor {
    t.Helper()
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
    if err != nil {
        t.Fatal(err)
    }
    info := &types.Info{
        Uses:  make(map[*ast.Ident]types.Object),
        Types: make(map[ast.Expr]types.TypeAndValue), // needed if strategy uses type info
    }
    conf := types.Config{Error: func(err error) {}}
    pkg, _ := conf.Check("sample", fset, []*ast.File{file}, info)

    ctx := &strategy.DiscoveryContext{
        Fset: fset, File: file, Pkg: pkg, Info: info,
        Src: []byte(src), FilePath: "test.go", PkgPath: "sample",
    }

    // Target node types for this strategy
    targetTypes := map[string]bool{
        "*ast.BinaryExpr": true,  // adjust to your strategy's NodeTypes()
    }

    var all []mutation.Descriptor
    for _, s := range strategy.All() {
        hasTarget := false
        for _, nt := range s.NodeTypes() {
            if targetTypes[nt] {
                hasTarget = true
                break
            }
        }
        if !hasTarget {
            continue
        }
        // IMPORTANT: walk once per strategy, not once per node type
        ast.Inspect(file, func(n ast.Node) bool {
            if n == nil {
                return false
            }
            switch n.(type) {
            case *ast.BinaryExpr:  // adjust to match targetTypes
                all = slices.AppendSeq(all, s.Discover(ctx, n))
            }
            return true
        })
    }
    return all
}
```

**Deduplication pitfall:** If your strategy registers multiple `NodeTypes()` (e.g. `["*ast.BlockStmt", "*ast.CaseClause"]`), the test helper must walk the AST **once per strategy**, not once per node type. Otherwise you get 2x the expected mutations. The pattern above handles this correctly by checking `hasTarget` with a `break`.

## Adding a Kind Constant

In `pkg/mutation/kind.go`, add your constant to the appropriate group:

```go
// Token-swap: group with similar kinds
const (
    MyKindFooToBar Kind = "mycategory/foo_to_bar"
    MyKindBarToFoo Kind = "mycategory/bar_to_foo"
)
```

Convention: `Kind = "category/name"` where:
- **category** = strategy package name (e.g. `arithmetic`, `chanops`, `lockremove`)
- **name** = snake_case describing the transformation

If the mutation may cause hangs (deadlocks, goroutine leaks), add it to `MayHang()`:

```go
func (k Kind) MayHang() bool {
    switch k {
    case LockRemoveUnlock, LockRemoveRUnlock,
        ChanOpsRemoveClose, ChanOpsRemoveSend, ChanOpsRemoveReceive,
        MyNewHangKind:  // add here
        return true
    }
    return false
}
```

## Registration

Add a blank import to `internal/strategies/all/all.go`:

```go
import (
    // ... existing imports ...
    _ "github.com/fredbi/go-mutesting/internal/strategies/mystrategy"
)
```

Keep imports sorted alphabetically.

## Key Design Constraints

1. **Strategies must be stateless** — no fields, no caches, pure functions of `(ctx, node)`
2. **Descriptors are value types** — no AST pointers, no closures, safe to serialize and pass across goroutines
3. **Iterators are lazy** — use `iter.Seq[mutation.Descriptor]` with `yield`; check `!yield(desc)` for early exit
4. **Use `ctx.Src` for human-readable `Original`** — `string(ctx.Src[startOffset:endOffset])`
5. **`ident.Obj == nil` distinguishes builtins** — for `close`, `len`, `cap`, etc. (no `Obj` means predeclared)
6. **Handle both `ExprStmt` and `DeferStmt`** — for statement-removing strategies (`defer close(ch)`, `defer mu.Unlock()`)

## Reference Files

| File | Purpose |
|:---|:---|
| `pkg/strategy/strategy.go` | `Strategy` interface, `DiscoveryContext` |
| `pkg/strategy/registry.go` | `Register()`, `All()`, `ForNodeType()` |
| `pkg/mutation/kind.go` | `Kind` constants, `MayHang()` |
| `pkg/mutation/descriptor.go` | `Descriptor`, `ApplySpec`, `StructuralAction` |
| `internal/strategies/arithmetic/` | Token-swap reference implementation |
| `internal/strategies/chanops/` | Structural (statement removal) reference implementation |
| `internal/strategies/slicebound/` | Structural (expression offset) reference implementation |
| `internal/strategies/argswap/` | Structural (type-aware, uses `Info.Types`) reference implementation |
| `internal/strategies/all/all.go` | Registration hub |
| `pkg/applier/structural.go` | Applier handlers for structural actions |
| `MUTATIONS.md` | Full mutation catalogue |
