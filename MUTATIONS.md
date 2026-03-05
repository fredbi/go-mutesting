# Mutation Catalogue

This document lists every mutation supported by the go-mutesting v2 pipeline.

Mutations are organized into two families:
- **Token-swap** mutations replace one token with another via byte-splicing (no AST re-parse needed).
- **Structural** mutations modify the AST (empty a block, remove a statement, replace an expression).

Each strategy self-registers via `init()` and is activated by blank-importing `internal/strategies/all`.

---

## Token-Swap Mutations

### Arithmetic

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `arithmetic/add_to_sub` | `+` | `-` | `arithmetic` | `a + b` &rarr; `a - b` |
| `arithmetic/sub_to_add` | `-` | `+` | `arithmetic` | `a - b` &rarr; `a + b` |
| `arithmetic/mul_to_div` | `*` | `/` | `arithmetic` | `a * b` &rarr; `a / b` |
| `arithmetic/div_to_mul` | `/` | `*` | `arithmetic` | `a / b` &rarr; `a * b` |
| `arithmetic/rem_to_mul` | `%` | `*` | `arithmetic` | `a % b` &rarr; `a * b` |

AST node: `*ast.BinaryExpr`

```go
// Original
func Total(price, tax int) int {
    return price + tax
}

// Mutated (arithmetic/add_to_sub)
func Total(price, tax int) int {
    return price - tax
}
```

---

### Conditional Boundary

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `conditional_boundary/less_to_less_eq` | `<` | `<=` | `conditional/boundary` | `a < b` &rarr; `a <= b` |
| `conditional_boundary/less_eq_to_less` | `<=` | `<` | `conditional/boundary` | `a <= b` &rarr; `a < b` |
| `conditional_boundary/greater_to_gr_eq` | `>` | `>=` | `conditional/boundary` | `a > b` &rarr; `a >= b` |
| `conditional_boundary/gr_eq_to_greater` | `>=` | `>` | `conditional/boundary` | `a >= b` &rarr; `a > b` |

AST node: `*ast.BinaryExpr`

```go
// Original
func IsPositive(n int) bool {
    return n > 0
}

// Mutated (conditional_boundary/greater_to_gr_eq)
func IsPositive(n int) bool {
    return n >= 0
}
```

---

### Conditional Negation

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `conditional_negation/eq_to_neq` | `==` | `!=` | `conditional/negation` | `a == b` &rarr; `a != b` |
| `conditional_negation/neq_to_eq` | `!=` | `==` | `conditional/negation` | `a != b` &rarr; `a == b` |
| `conditional_negation/less_to_gr_eq` | `<` | `>=` | `conditional/negation` | `a < b` &rarr; `a >= b` |
| `conditional_negation/gr_eq_to_less` | `>=` | `<` | `conditional/negation` | `a >= b` &rarr; `a < b` |
| `conditional_negation/greater_to_less_eq` | `>` | `<=` | `conditional/negation` | `a > b` &rarr; `a <= b` |
| `conditional_negation/less_eq_to_greater` | `<=` | `>` | `conditional/negation` | `a <= b` &rarr; `a > b` |

AST node: `*ast.BinaryExpr`

```go
// Original
func Contains(haystack []int, needle int) bool {
    for _, v := range haystack {
        if v == needle {
            return true
        }
    }
    return false
}

// Mutated (conditional_negation/eq_to_neq)
func Contains(haystack []int, needle int) bool {
    for _, v := range haystack {
        if v != needle {
            return true
        }
    }
    return false
}
```

---

### Logical

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `logical/and_to_or` | `&&` | `\|\|` | `logical` | `a && b` &rarr; `a \|\| b` |
| `logical/or_to_and` | `\|\|` | `&&` | `logical` | `a \|\| b` &rarr; `a && b` |

AST node: `*ast.BinaryExpr`

```go
// Original
func CanDrive(hasLicense, isAdult bool) bool {
    return hasLicense && isAdult
}

// Mutated (logical/and_to_or)
func CanDrive(hasLicense, isAdult bool) bool {
    return hasLicense || isAdult
}
```

---

### Bitwise

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `bitwise/and_to_or` | `&` | `\|` | `bitwise` | `a & b` &rarr; `a \| b` |
| `bitwise/or_to_and` | `\|` | `&` | `bitwise` | `a \| b` &rarr; `a & b` |
| `bitwise/xor_to_and` | `^` | `&` | `bitwise` | `a ^ b` &rarr; `a & b` |
| `bitwise/shl_to_shr` | `<<` | `>>` | `bitwise` | `a << b` &rarr; `a >> b` |
| `bitwise/shr_to_shl` | `>>` | `<<` | `bitwise` | `a >> b` &rarr; `a << b` |

AST node: `*ast.BinaryExpr`

```go
// Original
func Mask(flags, bit uint) uint {
    return flags & bit
}

// Mutated (bitwise/and_to_or)
func Mask(flags, bit uint) uint {
    return flags | bit
}
```

---

### Assignment Invert

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `assignment_invert/add_assign_to_sub_assign` | `+=` | `-=` | `assignment/invert` | `x += n` &rarr; `x -= n` |
| `assignment_invert/sub_assign_to_add_assign` | `-=` | `+=` | `assignment/invert` | `x -= n` &rarr; `x += n` |
| `assignment_invert/mul_assign_to_div_assign` | `*=` | `/=` | `assignment/invert` | `x *= n` &rarr; `x /= n` |
| `assignment_invert/div_assign_to_mul_assign` | `/=` | `*=` | `assignment/invert` | `x /= n` &rarr; `x *= n` |

AST node: `*ast.AssignStmt`

```go
// Original
func Accumulate(total *int, value int) {
    *total += value
}

// Mutated (assignment_invert/add_assign_to_sub_assign)
func Accumulate(total *int, value int) {
    *total -= value
}
```

---

### Assignment Remove (Self-Assignment)

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `assignment_remove/add_assign_to_assign` | `+=` | `=` | `assignment/remove_self` | `x += n` &rarr; `x = n` |
| `assignment_remove/sub_assign_to_assign` | `-=` | `=` | `assignment/remove_self` | `x -= n` &rarr; `x = n` |
| `assignment_remove/mul_assign_to_assign` | `*=` | `=` | `assignment/remove_self` | `x *= n` &rarr; `x = n` |
| `assignment_remove/div_assign_to_assign` | `/=` | `=` | `assignment/remove_self` | `x /= n` &rarr; `x = n` |
| `assignment_remove/rem_assign_to_assign` | `%=` | `=` | `assignment/remove_self` | `x %= n` &rarr; `x = n` |
| `assignment_remove/and_assign_to_assign` | `&=` | `=` | `assignment/remove_self` | `x &= n` &rarr; `x = n` |
| `assignment_remove/or_assign_to_assign` | `\|=` | `=` | `assignment/remove_self` | `x \|= n` &rarr; `x = n` |
| `assignment_remove/xor_assign_to_assign` | `^=` | `=` | `assignment/remove_self` | `x ^= n` &rarr; `x = n` |
| `assignment_remove/shl_assign_to_assign` | `<<=` | `=` | `assignment/remove_self` | `x <<= n` &rarr; `x = n` |
| `assignment_remove/shr_assign_to_assign` | `>>=` | `=` | `assignment/remove_self` | `x >>= n` &rarr; `x = n` |

AST node: `*ast.AssignStmt`

```go
// Original
func ApplyDiscount(price *float64, pct float64) {
    *price *= 1 - pct
}

// Mutated (assignment_remove/mul_assign_to_assign)
func ApplyDiscount(price *float64, pct float64) {
    *price = 1 - pct
}
```

---

### Bitwise Assignment

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `bitwise_assign/and_assign_to_or_assign` | `&=` | `\|=` | `bitwise_assign` | `x &= m` &rarr; `x \|= m` |
| `bitwise_assign/or_assign_to_and_assign` | `\|=` | `&=` | `bitwise_assign` | `x \|= m` &rarr; `x &= m` |
| `bitwise_assign/shl_assign_to_shr_assign` | `<<=` | `>>=` | `bitwise_assign` | `x <<= n` &rarr; `x >>= n` |
| `bitwise_assign/shr_assign_to_shl_assign` | `>>=` | `<<=` | `bitwise_assign` | `x >>= n` &rarr; `x <<= n` |

AST node: `*ast.AssignStmt`

```go
// Original
func ClearBit(flags *uint, bit uint) {
    *flags &= ^bit
}

// Mutated (bitwise_assign/and_assign_to_or_assign)
func ClearBit(flags *uint, bit uint) {
    *flags |= ^bit
}
```

---

### Increment / Decrement

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `incdec/inc_to_dec` | `++` | `--` | `incdec` | `i++` &rarr; `i--` |
| `incdec/dec_to_inc` | `--` | `++` | `incdec` | `i--` &rarr; `i++` |

AST node: `*ast.IncDecStmt`

```go
// Original
func CountUp(n *int) {
    *n++
}

// Mutated (incdec/inc_to_dec)
func CountUp(n *int) {
    *n--
}
```

---

### Loop Control

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `loopctrl/break_to_continue` | `break` | `continue` | `loopctrl` | `break` &rarr; `continue` |
| `loopctrl/continue_to_break` | `continue` | `break` | `loopctrl` | `continue` &rarr; `break` |

AST node: `*ast.BranchStmt` (only unlabeled)

```go
// Original
func FirstNegative(nums []int) int {
    for _, n := range nums {
        if n < 0 {
            return n
        }
        if n == 0 {
            continue
        }
    }
    return 0
}

// Mutated (loopctrl/continue_to_break)
func FirstNegative(nums []int) int {
    for _, n := range nums {
        if n < 0 {
            return n
        }
        if n == 0 {
            break
        }
    }
    return 0
}
```

---

### Negate Unary

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `negatives/remove_negation` | `-` (unary) | `+` (unary) | `negatives` | `-x` &rarr; `+x` |

AST node: `*ast.UnaryExpr`

```go
// Original
func Abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

// Mutated (negatives/remove_negation)
func Abs(x int) int {
    if x < 0 {
        return +x
    }
    return x
}
```

---

## Structural Mutations

### Branch: Empty If

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `branch/empty_if` | `branch/empty_if` | `*ast.IfStmt` |

Replaces the if-body with a type-aware noop (empty statement, or `_ = x` to keep variables used).

```go
// Original
func Clamp(x, lo, hi int) int {
    if x < lo {
        x = lo
    }
    return x
}

// Mutated (branch/empty_if)
func Clamp(x, lo, hi int) int {
    if x < lo {
        _ = lo
    }
    return x
}
```

---

### Branch: Empty Else

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `branch/empty_else` | `branch/empty_else` | `*ast.IfStmt` (with plain `else` block, not `else if`) |

Replaces the else-body with a type-aware noop.

```go
// Original
func Sign(n int) string {
    if n >= 0 {
        return "positive"
    } else {
        return "negative"
    }
}

// Mutated (branch/empty_else)
func Sign(n int) string {
    if n >= 0 {
        return "positive"
    } else {
    }
}
```

---

### Branch: Empty Case

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `branch/empty_case` | `branch/empty_case` | `*ast.CaseClause` (non-empty body) |

Replaces the case body with a type-aware noop.

```go
// Original
func Describe(n int) string {
    switch {
    case n > 0:
        return "positive"
    case n < 0:
        return "negative"
    default:
        return "zero"
    }
}

// Mutated (branch/empty_case — on the first case)
func Describe(n int) string {
    switch {
    case n > 0:
    case n < 0:
        return "negative"
    default:
        return "zero"
    }
}
```

---

### Statement Remove

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `statement/remove` | `statement/remove` | `*ast.BlockStmt`, `*ast.CaseClause` |

Replaces a single removable statement with a type-aware noop. Only targets:
- `AssignStmt` where the token is not `:=` (short variable declarations are not removable)
- `ExprStmt` (function calls used as statements)
- `IncDecStmt` (`i++`, `i--`)

```go
// Original
func Process(data []byte) {
    validate(data)
    count++
    result = transform(data)
}

// Mutated (statement/remove — removing "validate(data)")
func Process(data []byte) {
    _ = data
    count++
    result = transform(data)
}

// Mutated (statement/remove — removing "count++")
func Process(data []byte) {
    validate(data)
    _ = count
    result = transform(data)
}

// Mutated (statement/remove — removing "result = transform(data)")
func Process(data []byte) {
    validate(data)
    count++
    _ = data
}
```

---

### Expression: Remove Term

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `expression/remove_term` | `expression/remove_term` | `*ast.BinaryExpr` with `&&` or `\|\|` |

For `&&`: replaces one operand with `true`.
For `||`: replaces one operand with `false`.
Generates two mutations per expression (one for each operand).

```go
// Original
func IsEligible(age int, hasID bool) bool {
    return age >= 18 && hasID
}

// Mutated (expression/remove_term — left operand)
func IsEligible(age int, hasID bool) bool {
    return true && hasID
}

// Mutated (expression/remove_term — right operand)
func IsEligible(age int, hasID bool) bool {
    return age >= 18 && true
}
```

```go
// Original
func IsBlocked(banned, suspended bool) bool {
    return banned || suspended
}

// Mutated (expression/remove_term — left operand)
func IsBlocked(banned, suspended bool) bool {
    return false || suspended
}

// Mutated (expression/remove_term — right operand)
func IsBlocked(banned, suspended bool) bool {
    return banned || false
}
```

---

### Branch: Swap If/Else

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `branch/swap_if_else` | `branch/swap_if_else` | `*ast.IfStmt` (with plain `else` block, not `else if`) |

Swaps the if-body and else-body. Only targets `if/else` pairs where the else is a plain block (not an `else if` chain).

```go
// Original
func Compare(a, b int) bool {
    if a < b {
        return true
    } else {
        return false
    }
}

// Mutated (branch/swap_if_else)
func Compare(a, b int) bool {
    if a < b {
        return false
    } else {
        return true
    }
}
```

---

### Branch: Swap Case

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `branch/swap_case` | `branch/swap_case` | `*ast.SwitchStmt`, `*ast.TypeSwitchStmt` |

Swaps the bodies of adjacent case clauses. Generates one mutation per adjacent pair of non-empty cases.

```go
// Original
func Classify(x int) string {
    switch x {
    case 1:
        return "one"
    case 2:
        return "two"
    default:
        return "other"
    }
}

// Mutated (branch/swap_case — cases 0 and 1)
func Classify(x int) string {
    switch x {
    case 1:
        return "two"
    case 2:
        return "one"
    default:
        return "other"
    }
}

// Mutated (branch/swap_case — cases 1 and 2)
func Classify(x int) string {
    switch x {
    case 1:
        return "one"
    case 2:
        return "other"
    default:
        return "two"
    }
}
```

---

### Return: Nil Error

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `return/nil_error` | `return/nil_error` | `*ast.ReturnStmt` in functions returning `error` |

Replaces a non-nil error return value with `nil`. Requires the function signature to have an `error` result. Skips returns where the error is already `nil`.

```go
// Original
func Parse(s string) (int, error) {
    if s == "" {
        return 0, fmt.Errorf("empty input")
    }
    return len(s), nil
}

// Mutated (return/nil_error — first return only, second already returns nil)
func Parse(s string) (int, error) {
    if s == "" {
        return 0, nil
    }
    return len(s), nil
}
```

---

### Return: Zero Value

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `return/zero_value` | `return/zero_value` | `*ast.ReturnStmt` in functions returning non-error values |

Replaces a non-error return value with its zero value (`0` for ints, `""` for strings, `nil` for pointers/slices/maps, `false` for bools). Skips values that are already zero and skips `error`-typed results (those are handled by `return/nil_error`).

```go
// Original
func Lookup(key string) (string, bool) {
    if key == "found" {
        return "value", true
    }
    return "", false
}

// Mutated (return/zero_value — zeroing the string in first return)
func Lookup(key string) (string, bool) {
    if key == "found" {
        return "", true
    }
    return "", false
}
```

```go
// Original
func GetItems() []int {
    return []int{1, 2, 3}
}

// Mutated (return/zero_value)
func GetItems() []int {
    return nil
}
```

```go
// Original
func GetPtr() *Config {
    return &Config{Port: 8080}
}

// Mutated (return/zero_value)
func GetPtr() *Config {
    return nil
}
```

---

### Return: Negate Bool

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `return/negate_bool` | `return/negate_bool` | `*ast.ReturnStmt` in functions returning `bool` |

Negates a boolean return value by wrapping it with `!`. Skips expressions that are already negated (to avoid `!!x`). Works for multi-return functions too (only the `bool` position is negated).

```go
// Original
func IsValid(s string) bool {
    return len(s) > 0
}

// Mutated (return/negate_bool)
func IsValid(s string) bool {
    return !(len(s) > 0)
}
```

```go
// Original
func Lookup(key string) (string, bool) {
    if key == "found" {
        return "value", true
    }
    return "", false
}

// Mutated (return/negate_bool — first return)
func Lookup(key string) (string, bool) {
    if key == "found" {
        return "value", !true
    }
    return "", false
}

// Mutated (return/negate_bool — second return)
func Lookup(key string) (string, bool) {
    if key == "found" {
        return "value", true
    }
    return "", !false
}
```

---

## Literal Mutations

### Boolean Literal

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `bool_literal/true_to_false` | `true` | `false` | `boolliteral` | `true` &rarr; `false` |
| `bool_literal/false_to_true` | `false` | `true` | `boolliteral` | `false` &rarr; `true` |

AST node: `*ast.Ident` (only built-in `true`/`false`, not user-defined identifiers)

```go
// Original
var debug = true
var verbose = false

func Check() bool {
    return true
}

// Mutated (bool_literal/true_to_false — var debug)
var debug = false

// Mutated (bool_literal/false_to_true — var verbose)
var verbose = true

// Mutated (bool_literal/true_to_false — return)
func Check() bool {
    return false
}
```

---

### Conditional Expression

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `conditional_expr/replace_with_true` | condition | `true` | `condexpr` | `x > 0` &rarr; `true` |
| `conditional_expr/replace_with_false` | condition | `false` | `condexpr` | `x > 0` &rarr; `false` |

AST node: `*ast.IfStmt`, `*ast.ForStmt`

Replaces the condition of `if` and `for` statements. For `for` loops, only `replace_with_false` is generated (replacing with `true` would create an infinite loop). Skips conditions that are already literal `true` or `false`.

```go
// Original
func Check(x int) string {
    if x > 0 {
        return "positive"
    }
    return "non-positive"
}

// Mutated (conditional_expr/replace_with_true)
func Check(x int) string {
    if true {
        return "positive"
    }
    return "non-positive"
}

// Mutated (conditional_expr/replace_with_false)
func Check(x int) string {
    if false {
        return "positive"
    }
    return "non-positive"
}
```

```go
// Original
func Counter(n int) int {
    total := 0
    for i := 0; i < n; i++ {
        total += i
    }
    return total
}

// Mutated (conditional_expr/replace_with_false — for loop, no true variant)
func Counter(n int) int {
    total := 0
    for i := 0; false; i++ {
        total += i
    }
    return total
}
```

---

### String Literal

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `string_literal/non_empty_to_empty` | `"hello"` | `""` | `stringliteral` | `"hello"` &rarr; `""` |
| `string_literal/empty_to_sentinel` | `""` | `"MUTATED"` | `stringliteral` | `""` &rarr; `"MUTATED"` |

AST node: `*ast.BasicLit` (only interpreted string literals, not raw strings or non-string literals)

Deliberately skips raw strings (backtick-quoted) to avoid breaking regex patterns and multi-line content. Only targets `token.STRING` literals that start with `"`.

```go
// Original
var greeting = "hello"
var empty = ""

// Mutated (string_literal/non_empty_to_empty)
var greeting = ""

// Mutated (string_literal/empty_to_sentinel)
var empty = "MUTATED"
```

---

## Go-Idiomatic Mutations

### Argument Swap

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `argswap/swap_arguments` | `argswap` | `*ast.CallExpr` |

Swaps adjacent function call arguments that have the same type. Requires type information (`types.Info.Types`) to determine argument types. Skips:
- Calls with fewer than 2 arguments
- Variadic expansion calls (`f(slice...)`)
- Argument pairs with different types
- Argument pairs with identical source text (equivalent mutants)

Works for regular function calls, method calls, and built-in function calls.

```go
// Original
func FullName(first, last string) string {
    return Concat(first, last)
}

// Mutated (argswap/swap_arguments)
func FullName(first, last string) string {
    return Concat(last, first)
}
```

```go
// Original — three same-type args produce 2 mutations (adjacent pairs)
func use() int {
    return triple(1, 2, 3)
}

// Mutated (argswap/swap_arguments — args 0,1)
func use() int {
    return triple(2, 1, 3)
}

// Mutated (argswap/swap_arguments — args 1,2)
func use() int {
    return triple(1, 3, 2)
}
```

```go
// Original — mixed types: only same-type adjacent pair is swapped
func use() {
    f(1, 2, true)  // int, int, bool
}

// Mutated (argswap/swap_arguments — only the two ints)
func use() {
    f(2, 1, true)
}
```

---

### Panic to Return

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `panic/replace_with_return` | `panicremove` | `*ast.ExprStmt` (calls to built-in `panic`) |

Replaces `panic(...)` calls with `return` statements using zero values matching the enclosing function's return signature. Uses type information to ensure only calls to the built-in `panic` are targeted (not user-defined functions named "panic").

For void functions, generates bare `return`. For functions with return values, generates `return` with appropriate zero values (`0`, `""`, `nil`, `false`, etc.).

```go
// Original
func MustParse(s string) int {
    if s == "" {
        panic("empty input")
    }
    return len(s)
}

// Mutated (panic/replace_with_return)
func MustParse(s string) int {
    if s == "" {
        return 0
    }
    return len(s)
}
```

```go
// Original
func MustLookup(key string) (string, error) {
    panic("not implemented")
}

// Mutated (panic/replace_with_return)
func MustLookup(key string) (string, error) {
    return "", nil
}
```

```go
// Original
func PanicVoid() {
    panic("fatal")
}

// Mutated (panic/replace_with_return)
func PanicVoid() {
    return
}
```

---

## Severity Mutations — Panic Risk

### Nil Guard Removal

| Kind | Strategy | Targets |
|:-----|:---------|:--------|
| `nilguard/remove_nil_check` | `nilguard` | `*ast.IfStmt` with `x != nil` or `x == nil` conditions |

Removes nil safety checks by replacing the condition with the value that bypasses the guard:
- `if x != nil { use(x) }` — condition becomes `true` (body always runs, even when x is nil)
- `if x == nil { return }` — condition becomes `false` (guard never triggers, falls through to use nil)

```go
// Original
func SafeLen(s *string) int {
    if s != nil {
        return len(*s)
    }
    return 0
}

// Mutated (nilguard/remove_nil_check)
func SafeLen(s *string) int {
    if true {
        return len(*s)  // panics when s is nil
    }
    return 0
}
```

---

### Slice Bound Off-By-One

| Kind | Original | Mutated | Strategy | Example |
|:-----|:---------|:--------|:---------|:--------|
| `slicebound/index_plus_one` | `s[i]` | `s[i+1]` | `slicebound` | `items[0]` &rarr; `items[0+1]` |
| `slicebound/index_minus_one` | `s[i]` | `s[i-1]` | `slicebound` | `items[i]` &rarr; `items[i-1]` |
| `slicebound/slice_high_plus_one` | `s[:n]` | `s[:n+1]` | `slicebound` | `items[:n]` &rarr; `items[:n+1]` |
| `slicebound/slice_low_plus_one` | `s[n:]` | `s[n+1:]` | `slicebound` | `items[n:]` &rarr; `items[n+1:]` |

AST nodes: `*ast.IndexExpr`, `*ast.SliceExpr`

Introduces off-by-one errors that typically cause panics (index out of range, slice bounds out of range).

```go
// Original
func Head(items []int) int {
    return items[0]
}

// Mutated (slicebound/index_plus_one)
func Head(items []int) int {
    return items[0 + 1]  // panics for single-element slices
}
```

```go
// Original
func Take(items []int, n int) []int {
    return items[:n]
}

// Mutated (slicebound/slice_high_plus_one)
func Take(items []int, n int) []int {
    return items[:n + 1]  // panics when n == len(items)
}
```

---

## Severity Mutations — Hang Risk

These mutations may cause deadlocks or goroutine leaks. The execution engine should use aggressive timeouts. `Kind.MayHang()` returns `true` for the hang-risk kinds.

### Lock/Unlock Removal

| Kind | Original | Effect | MayHang |
|:-----|:---------|:-------|:--------|
| `lockremove/remove_lock` | `mu.Lock()` | Subsequent Unlock panics (unlock of unlocked mutex) | No |
| `lockremove/remove_unlock` | `mu.Unlock()` | Next Lock deadlocks | **Yes** |
| `lockremove/remove_rlock` | `mu.RLock()` | Subsequent RUnlock panics | No |
| `lockremove/remove_runlock` | `mu.RUnlock()` | Next RLock may deadlock | **Yes** |

AST nodes: `*ast.BlockStmt`, `*ast.CaseClause` (iterates over statements)

Targets calls by method name (Lock, Unlock, RLock, RUnlock) with zero arguments. Also targets `defer` statements (`defer mu.Unlock()` is idiomatic Go).

```go
// Original
func SafeIncrement(mu *sync.Mutex, counter *int) {
    mu.Lock()
    defer mu.Unlock()
    *counter++
}

// Mutated (lockremove/remove_lock — panic: unlock of unlocked mutex)
func SafeIncrement(mu *sync.Mutex, counter *int) {
    _ = mu          // noop replaces mu.Lock()
    defer mu.Unlock()
    *counter++
}

// Mutated (lockremove/remove_unlock — HANG: next Lock deadlocks)
func SafeIncrement(mu *sync.Mutex, counter *int) {
    mu.Lock()
    _ = mu          // noop replaces defer mu.Unlock()
    *counter++
}
```

---

### Channel Operation Removal

| Kind | Original | Effect | MayHang |
|:-----|:---------|:-------|:--------|
| `chanops/remove_close` | `close(ch)` | Range-over-channel readers block forever | **Yes** |
| `chanops/remove_send` | `ch <- val` | Receivers block forever | **Yes** |
| `chanops/remove_receive` | `<-ch` | Senders may block (unbuffered channels) | **Yes** |

AST nodes: `*ast.BlockStmt`, `*ast.CaseClause` (iterates over statements)

Targets channel operations: `close(ch)` (builtin only), `ch <- val`, `<-ch` (standalone). Also targets `defer close(ch)`.

```go
// Original
func Producer(ch chan int) {
    ch <- 1
    ch <- 2
    close(ch)
}

// Mutated (chanops/remove_close — HANG: range readers block)
func Producer(ch chan int) {
    ch <- 1
    ch <- 2
    _ = ch          // noop replaces close(ch)
}

// Mutated (chanops/remove_send — HANG: receivers block)
func Producer(ch chan int) {
    _ = ch          // noop replaces ch <- 1
    ch <- 2
    close(ch)
}
```

---

## Summary

| Family | Category | Strategy | Mutations | AST Node |
|:-------|:---------|:---------|----------:|:---------|
| Token-swap | Arithmetic | `arithmetic` | 5 | `*ast.BinaryExpr` |
| Token-swap | Conditional Boundary | `conditional/boundary` | 4 | `*ast.BinaryExpr` |
| Token-swap | Conditional Negation | `conditional/negation` | 6 | `*ast.BinaryExpr` |
| Token-swap | Logical | `logical` | 2 | `*ast.BinaryExpr` |
| Token-swap | Bitwise | `bitwise` | 5 | `*ast.BinaryExpr` |
| Token-swap | Assignment Invert | `assignment/invert` | 4 | `*ast.AssignStmt` |
| Token-swap | Assignment Remove | `assignment/remove_self` | 10 | `*ast.AssignStmt` |
| Token-swap | Bitwise Assignment | `bitwise_assign` | 4 | `*ast.AssignStmt` |
| Token-swap | Increment/Decrement | `incdec` | 2 | `*ast.IncDecStmt` |
| Token-swap | Loop Control | `loopctrl` | 2 | `*ast.BranchStmt` |
| Token-swap | Negate Unary | `negatives` | 1 | `*ast.UnaryExpr` |
| Structural | Branch Empty If | `branch/empty_if` | 1 | `*ast.IfStmt` |
| Structural | Branch Empty Else | `branch/empty_else` | 1 | `*ast.IfStmt` |
| Structural | Branch Empty Case | `branch/empty_case` | 1 | `*ast.CaseClause` |
| Structural | Branch Swap If/Else | `branch/swap_if_else` | 1 | `*ast.IfStmt` |
| Structural | Branch Swap Case | `branch/swap_case` | 1 | `*ast.SwitchStmt`, `*ast.TypeSwitchStmt` |
| Structural | Statement Remove | `statement/remove` | 1 | `*ast.BlockStmt`, `*ast.CaseClause` |
| Structural | Expression Remove Term | `expression/remove_term` | 2 | `*ast.BinaryExpr` |
| Structural | Return Nil Error | `return/nil_error` | 1 | `*ast.ReturnStmt` |
| Structural | Return Zero Value | `return/zero_value` | 1 | `*ast.ReturnStmt` |
| Structural | Return Negate Bool | `return/negate_bool` | 1 | `*ast.ReturnStmt` |
| Literal | Boolean Literal | `boolliteral` | 2 | `*ast.Ident` |
| Literal | Conditional Expression | `condexpr` | 2 | `*ast.IfStmt`, `*ast.ForStmt` |
| Literal | String Literal | `stringliteral` | 2 | `*ast.BasicLit` |
| Go-Idiomatic | Argument Swap | `argswap` | 1 | `*ast.CallExpr` |
| Go-Idiomatic | Panic to Return | `panicremove` | 1 | `*ast.ExprStmt` |
| Severity (panic) | Nil Guard Removal | `nilguard` | 1 | `*ast.IfStmt` |
| Severity (panic) | Slice Bound | `slicebound` | 4 | `*ast.IndexExpr`, `*ast.SliceExpr` |
| Severity (hang) | Lock/Unlock Removal | `lockremove` | 4 | `*ast.BlockStmt`, `*ast.CaseClause` |
| Severity (hang) | Channel Op Removal | `chanops` | 3 | `*ast.BlockStmt`, `*ast.CaseClause` |
| | | **Total distinct rules** | **76** | |
