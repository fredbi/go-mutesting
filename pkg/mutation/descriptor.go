package mutation

// Position represents a source code position.
type Position struct {
	Line   int
	Column int
	Offset int // byte offset from start of file
}

// ApplySpec is a tagged union that tells the Applier how to apply a mutation.
// Exactly one field should be non-nil.
type ApplySpec struct {
	TokenSwap  *TokenSwapSpec
	Structural *StructuralSpec
}

// TokenSwapSpec describes a token-level mutation (byte splice).
type TokenSwapSpec struct {
	OriginalToken    string
	ReplacementToken string
	StartOffset      int // byte offset in file
	EndOffset        int // byte offset in file (exclusive)
}

// StructuralAction describes the kind of structural mutation to apply.
type StructuralAction int

const (
	ActionEmptyBlock       StructuralAction = iota // Replace block body with noop
	ActionRemoveStatement                          // Replace a single statement with noop
	ActionReplaceWithTrue                          // Replace expression with "true"
	ActionReplaceWithFalse                         // Replace expression with "false"
	ActionSwapIfElse                               // Swap if-body and else-body
	ActionSwapCase                                 // Swap two adjacent case clause bodies
	ActionReturnZero                               // Replace return values with zero values
	ActionNegateBoolReturn                         // Negate boolean return value
	ActionReplaceStmtWithReturn                    // Replace a statement (e.g. panic) with a return
	ActionSwapCallArgs                             // Swap two arguments in a function call
	ActionOffsetExpr                               // Offset an expression by +N or -N (e.g. s[i] → s[i+1])
)

func (a StructuralAction) String() string {
	switch a {
	case ActionEmptyBlock:
		return "empty_block"
	case ActionRemoveStatement:
		return "remove_statement"
	case ActionReplaceWithTrue:
		return "replace_with_true"
	case ActionReplaceWithFalse:
		return "replace_with_false"
	case ActionSwapIfElse:
		return "swap_if_else"
	case ActionSwapCase:
		return "swap_case"
	case ActionReturnZero:
		return "return_zero"
	case ActionNegateBoolReturn:
		return "negate_bool_return"
	case ActionReplaceStmtWithReturn:
		return "replace_stmt_with_return"
	case ActionSwapCallArgs:
		return "swap_call_args"
	case ActionOffsetExpr:
		return "offset_expr"
	default:
		return "unknown"
	}
}

// StructuralSpec describes an AST-structural mutation.
type StructuralSpec struct {
	NodeType     string           // e.g. "IfStmt", "BlockStmt", "CaseClause", "BinaryExpr"
	Action       StructuralAction
	TargetIndex  int // which statement/clause in a list (-1 if N/A)
	TargetIndex2 int // second index for swap mutations (-1 if N/A)
	StartOffset  int // byte offset for the target node/statement
	EndOffset    int // byte offset (exclusive)

	// ReturnMeta carries metadata for return-value mutations.
	// Each entry describes the replacement for one return value.
	// Empty string means "keep original". Non-empty means "replace with this expression".
	ReturnMeta []string
}

// Descriptor is a plain value type describing a single mutation.
// It contains no AST references, no closures, and is safe to serialize,
// compare, deduplicate, and pass across goroutines.
type Descriptor struct {
	ID          string   // Stable hash: SHA-256(File, Kind, Line, Col, Original, Replacement)[:32]
	File        string   // Absolute path
	PkgPath     string   // Go import path
	FuncName    string   // Enclosing function name (for filtering)
	StartPos    Position // Start of mutated region
	EndPos      Position // End of mutated region
	Kind        Kind
	Status      Status
	Original    string // Human-readable original (e.g. "+", "if body: 3 statements")
	Replacement string // Human-readable replacement (e.g. "-", "empty block")
	Apply       ApplySpec
}
