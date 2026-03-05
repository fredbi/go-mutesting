package mutation

import "strings"

// Kind is a hierarchical mutation kind string in the form "category/name".
type Kind string

// Category returns the category part of the kind (before the '/').
func (k Kind) Category() string {
	if i := strings.IndexByte(string(k), '/'); i >= 0 {
		return string(k[:i])
	}
	return string(k)
}

// Name returns the name part of the kind (after the '/').
func (k Kind) Name() string {
	if i := strings.IndexByte(string(k), '/'); i >= 0 {
		return string(k[i+1:])
	}
	return ""
}

// Token-swap mutation kinds (ported from gremlins).
const (
	ArithmeticAddToSub Kind = "arithmetic/add_to_sub"
	ArithmeticSubToAdd Kind = "arithmetic/sub_to_add"
	ArithmeticMulToDiv Kind = "arithmetic/mul_to_div"
	ArithmeticDivToMul Kind = "arithmetic/div_to_mul"
	ArithmeticRemToMul Kind = "arithmetic/rem_to_mul"

	ConditionalBoundaryLessToLessEq    Kind = "conditional_boundary/less_to_less_eq"
	ConditionalBoundaryLessEqToLess    Kind = "conditional_boundary/less_eq_to_less"
	ConditionalBoundaryGreaterToGrEq   Kind = "conditional_boundary/greater_to_gr_eq"
	ConditionalBoundaryGrEqToGreater   Kind = "conditional_boundary/gr_eq_to_greater"
	ConditionalNegationEqToNeq         Kind = "conditional_negation/eq_to_neq"
	ConditionalNegationNeqToEq         Kind = "conditional_negation/neq_to_eq"
	ConditionalNegationLessToGrEq      Kind = "conditional_negation/less_to_gr_eq"
	ConditionalNegationGrEqToLess      Kind = "conditional_negation/gr_eq_to_less"
	ConditionalNegationGreaterToLessEq Kind = "conditional_negation/greater_to_less_eq"
	ConditionalNegationLessEqToGreater Kind = "conditional_negation/less_eq_to_greater"

	LogicalAndToOr Kind = "logical/and_to_or"
	LogicalOrToAnd Kind = "logical/or_to_and"

	BitwiseAndToOr  Kind = "bitwise/and_to_or"
	BitwiseOrToAnd  Kind = "bitwise/or_to_and"
	BitwiseXorToAnd Kind = "bitwise/xor_to_and"
	BitwiseShlToShr Kind = "bitwise/shl_to_shr"
	BitwiseShrToShl Kind = "bitwise/shr_to_shl"

	AssignmentInvertAddAssign Kind = "assignment_invert/add_assign_to_sub_assign"
	AssignmentInvertSubAssign Kind = "assignment_invert/sub_assign_to_add_assign"
	AssignmentInvertMulAssign Kind = "assignment_invert/mul_assign_to_div_assign"
	AssignmentInvertDivAssign Kind = "assignment_invert/div_assign_to_mul_assign"
	AssignmentRemoveAdd       Kind = "assignment_remove/add_assign_to_assign"
	AssignmentRemoveSub       Kind = "assignment_remove/sub_assign_to_assign"
	AssignmentRemoveMul       Kind = "assignment_remove/mul_assign_to_assign"
	AssignmentRemoveDiv       Kind = "assignment_remove/div_assign_to_assign"
	AssignmentRemoveRem       Kind = "assignment_remove/rem_assign_to_assign"
	AssignmentRemoveAndAssign Kind = "assignment_remove/and_assign_to_assign"
	AssignmentRemoveOrAssign  Kind = "assignment_remove/or_assign_to_assign"
	AssignmentRemoveXorAssign Kind = "assignment_remove/xor_assign_to_assign"
	AssignmentRemoveShlAssign Kind = "assignment_remove/shl_assign_to_assign"
	AssignmentRemoveShrAssign Kind = "assignment_remove/shr_assign_to_assign"

	BitwiseAssignAndToOr  Kind = "bitwise_assign/and_assign_to_or_assign"
	BitwiseAssignOrToAnd  Kind = "bitwise_assign/or_assign_to_and_assign"
	BitwiseAssignShlToShr Kind = "bitwise_assign/shl_assign_to_shr_assign"
	BitwiseAssignShrToShl Kind = "bitwise_assign/shr_assign_to_shl_assign"

	IncDecIncToDec Kind = "incdec/inc_to_dec"
	IncDecDecToInc Kind = "incdec/dec_to_inc"

	LoopCtrlBreakToContinue Kind = "loopctrl/break_to_continue"
	LoopCtrlContinueToBreak Kind = "loopctrl/continue_to_break"

	NegativesRemove Kind = "negatives/remove_negation"
)

// Structural mutation kinds (ported from go-mutesting).
const (
	BranchEmptyIf   Kind = "branch/empty_if"
	BranchEmptyElse Kind = "branch/empty_else"
	BranchEmptyCase Kind = "branch/empty_case"

	StatementRemove Kind = "statement/remove"

	ExpressionRemoveTerm Kind = "expression/remove_term"
)

// New structural mutation kinds.
const (
	BranchSwapIfElse Kind = "branch/swap_if_else"
	BranchSwapCase   Kind = "branch/swap_case"

	ReturnNilError   Kind = "return/nil_error"
	ReturnZeroValue  Kind = "return/zero_value"
	ReturnNegateBool Kind = "return/negate_bool"
)

// Literal mutation kinds (inspired by stryker4s).
const (
	BoolLitTrueToFalse Kind = "bool_literal/true_to_false"
	BoolLitFalseToTrue Kind = "bool_literal/false_to_true"

	ConditionalExprTrue  Kind = "conditional_expr/replace_with_true"
	ConditionalExprFalse Kind = "conditional_expr/replace_with_false"

	StringLitNonEmptyToEmpty Kind = "string_literal/non_empty_to_empty"
	StringLitEmptyToSentinel Kind = "string_literal/empty_to_sentinel"
)

// Go-idiomatic mutation kinds.
const (
	PanicToReturn Kind = "panic/replace_with_return"
	ArgSwap       Kind = "argswap/swap_arguments"
)

// Severity mutation kinds — may cause panics (nil dereference, index out of range).
const (
	NilGuardRemove Kind = "nilguard/remove_nil_check"

	SliceBoundIndexUp  Kind = "slicebound/index_plus_one"
	SliceBoundIndexDown Kind = "slicebound/index_minus_one"
	SliceBoundHighUp   Kind = "slicebound/slice_high_plus_one"
	SliceBoundLowUp    Kind = "slicebound/slice_low_plus_one"
)

// Severity mutation kinds — may cause hangs (deadlocks, goroutine leaks).
// The execution engine should use aggressive timeouts for these.
const (
	LockRemoveLock    Kind = "lockremove/remove_lock"
	LockRemoveUnlock  Kind = "lockremove/remove_unlock"
	LockRemoveRLock   Kind = "lockremove/remove_rlock"
	LockRemoveRUnlock Kind = "lockremove/remove_runlock"

	ChanOpsRemoveClose   Kind = "chanops/remove_close"
	ChanOpsRemoveSend    Kind = "chanops/remove_send"
	ChanOpsRemoveReceive Kind = "chanops/remove_receive"
)

// MayHang reports whether this mutation kind is likely to cause hangs
// (deadlocks, goroutine leaks). The execution engine should use
// aggressive timeouts for such mutations.
func (k Kind) MayHang() bool {
	switch k {
	case LockRemoveUnlock, LockRemoveRUnlock,
		ChanOpsRemoveClose, ChanOpsRemoveSend, ChanOpsRemoveReceive:
		return true
	}
	return false
}
