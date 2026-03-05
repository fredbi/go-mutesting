package stringliteral

import (
	"go/ast"
	"go/token"
	"iter"
	"strings"

	"github.com/fredbi/go-mutesting/pkg/mutation"
	"github.com/fredbi/go-mutesting/pkg/strategy"
)

func init() {
	strategy.Register(&stringLiteralStrategy{})
}

// SentinelString is the replacement for empty strings, following stryker4s convention.
const SentinelString = "Stryker was here!"

type stringLiteralStrategy struct{}

func (s *stringLiteralStrategy) Name() string       { return "string_literal" }
func (s *stringLiteralStrategy) NodeTypes() []string { return []string{"*ast.BasicLit"} }

func (s *stringLiteralStrategy) Discover(ctx *strategy.DiscoveryContext, node ast.Node) iter.Seq[mutation.Descriptor] {
	return func(yield func(mutation.Descriptor) bool) {
		lit, ok := node.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return
		}

		pos := ctx.Fset.Position(lit.Pos())
		endOffset := pos.Offset + len(lit.Value)

		// Determine if this is an empty string or non-empty string.
		// lit.Value includes quotes, e.g. `""`, `"hello"`, "`hello`".
		value := lit.Value

		if isEmptyString(value) {
			// Empty -> sentinel.
			replacement := `"` + SentinelString + `"`

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: pos.Line, Column: pos.Column, Offset: pos.Offset,
				},
				EndPos: mutation.Position{
					Line: pos.Line, Column: pos.Column + len(value), Offset: endOffset,
				},
				Kind:        mutation.StringLitEmptyToSentinel,
				Status:      mutation.Runnable,
				Original:    value,
				Replacement: replacement,
				Apply: mutation.ApplySpec{
					TokenSwap: &mutation.TokenSwapSpec{
						OriginalToken:    value,
						ReplacementToken: replacement,
						StartOffset:      pos.Offset,
						EndOffset:        endOffset,
					},
				},
			}

			yield(desc)
		} else if isSimpleString(value) {
			// Non-empty -> empty.
			replacement := `""`

			desc := mutation.Descriptor{
				File:    ctx.FilePath,
				PkgPath: ctx.PkgPath,
				StartPos: mutation.Position{
					Line: pos.Line, Column: pos.Column, Offset: pos.Offset,
				},
				EndPos: mutation.Position{
					Line: pos.Line, Column: pos.Column + len(value), Offset: endOffset,
				},
				Kind:        mutation.StringLitNonEmptyToEmpty,
				Status:      mutation.Runnable,
				Original:    value,
				Replacement: replacement,
				Apply: mutation.ApplySpec{
					TokenSwap: &mutation.TokenSwapSpec{
						OriginalToken:    value,
						ReplacementToken: replacement,
						StartOffset:      pos.Offset,
						EndOffset:        endOffset,
					},
				},
			}

			yield(desc)
		}
	}
}

// isEmptyString checks if a Go string literal represents an empty string.
func isEmptyString(raw string) bool {
	return raw == `""` || raw == "``"
}

// isSimpleString checks if a Go string literal is a non-empty regular double-quoted string.
// We skip raw strings (backtick) since they may contain multi-line content, regex patterns, etc.
// that would break if replaced with "".
func isSimpleString(raw string) bool {
	return len(raw) > 2 && strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`)
}
