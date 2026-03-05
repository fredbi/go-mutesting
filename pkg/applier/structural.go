package applier

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"github.com/fredbi/go-mutesting/pkg/mutation"
)

// StructuralApplier applies AST-structural mutations.
// It re-parses the file from the workdir, locates the target node
// by byte offset, applies the structural action, and writes back.
type StructuralApplier struct{}

func (a *StructuralApplier) Apply(desc mutation.Descriptor, workdir string) error {
	spec := desc.Apply.Structural
	if spec == nil {
		return fmt.Errorf("descriptor %s: no StructuralSpec", desc.ID)
	}

	target := filepath.Join(workdir, desc.File)
	src, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("reading %s: %w", target, err)
	}

	// Save original for rollback.
	backup := filepath.Join(workdir, desc.File+".orig")
	if err := os.WriteFile(backup, src, 0o644); err != nil {
		return fmt.Errorf("writing backup: %w", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, target, src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", target, err)
	}

	// Type-check for noop generation.
	pkg, info := typeCheck(fset, file)

	if err := applyStructural(fset, file, pkg, info, spec); err != nil {
		return fmt.Errorf("applying mutation %s: %w", desc.ID, err)
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return fmt.Errorf("printing %s: %w", desc.ID, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting %s: %w", desc.ID, err)
	}

	return os.WriteFile(target, formatted, 0o644)
}

func (a *StructuralApplier) Rollback(desc mutation.Descriptor, workdir string) error {
	target := filepath.Join(workdir, desc.File)
	backup := filepath.Join(workdir, desc.File+".orig")

	data, err := os.ReadFile(backup)
	if err != nil {
		return fmt.Errorf("reading backup for rollback: %w", err)
	}

	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("restoring original: %w", err)
	}

	return os.Remove(backup)
}

func applyStructural(fset *token.FileSet, file *ast.File, pkg *types.Package, info *types.Info, spec *mutation.StructuralSpec) error {
	switch spec.Action {
	case mutation.ActionEmptyBlock:
		return applyEmptyBlock(fset, file, pkg, info, spec)
	case mutation.ActionRemoveStatement:
		return applyRemoveStatement(fset, file, pkg, info, spec)
	case mutation.ActionReplaceWithTrue:
		return applyReplaceExpr(fset, file, spec, "true")
	case mutation.ActionReplaceWithFalse:
		return applyReplaceExpr(fset, file, spec, "false")
	case mutation.ActionSwapIfElse:
		return applySwapIfElse(fset, file, spec)
	case mutation.ActionSwapCase:
		return applySwapCase(fset, file, spec)
	case mutation.ActionReturnZero:
		return applyReturnReplace(fset, file, spec)
	case mutation.ActionNegateBoolReturn:
		return applyReturnNegate(fset, file, spec)
	case mutation.ActionReplaceStmtWithReturn:
		return applyReplaceStmtWithReturn(fset, file, spec)
	case mutation.ActionSwapCallArgs:
		return applySwapCallArgs(fset, file, spec)
	case mutation.ActionOffsetExpr:
		return applyOffsetExpr(fset, file, spec)
	default:
		return fmt.Errorf("unknown structural action: %v", spec.Action)
	}
}

func applyEmptyBlock(fset *token.FileSet, file *ast.File, pkg *types.Package, info *types.Info, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		switch spec.NodeType {
		case "IfStmt":
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok {
				return true
			}
			pos := fset.Position(ifStmt.Body.Lbrace)
			if pos.Offset == spec.StartOffset {
				noop := createNoopOfStatements(pkg, info, ifStmt.Body.List)
				ifStmt.Body.List = []ast.Stmt{noop}
				applied = true
				return false
			}

		case "IfStmt.Else":
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok {
				return true
			}
			block, ok := ifStmt.Else.(*ast.BlockStmt)
			if !ok {
				return true
			}
			pos := fset.Position(block.Lbrace)
			if pos.Offset == spec.StartOffset {
				noop := createNoopOfStatements(pkg, info, block.List)
				block.List = []ast.Stmt{noop}
				applied = true
				return false
			}

		case "CaseClause":
			cc, ok := n.(*ast.CaseClause)
			if !ok {
				return true
			}
			pos := fset.Position(cc.Colon)
			if pos.Offset == spec.StartOffset {
				noop := createNoopOfStatements(pkg, info, cc.Body)
				cc.Body = []ast.Stmt{noop}
				applied = true
				return false
			}
		}

		return true
	})

	if !applied {
		return fmt.Errorf("could not find target node at offset %d for %s", spec.StartOffset, spec.NodeType)
	}
	return nil
}

func applyRemoveStatement(fset *token.FileSet, file *ast.File, pkg *types.Package, info *types.Info, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		var stmts []ast.Stmt
		var setStmt func(int, ast.Stmt)

		switch node := n.(type) {
		case *ast.BlockStmt:
			stmts = node.List
			setStmt = func(i int, s ast.Stmt) { node.List[i] = s }
		case *ast.CaseClause:
			stmts = node.Body
			setStmt = func(i int, s ast.Stmt) { node.Body[i] = s }
		default:
			return true
		}

		if spec.TargetIndex < 0 || spec.TargetIndex >= len(stmts) {
			return true
		}

		stmt := stmts[spec.TargetIndex]
		pos := fset.Position(stmt.Pos())
		if pos.Offset == spec.StartOffset {
			noop := createNoopOfStatement(pkg, info, stmt)
			setStmt(spec.TargetIndex, noop)
			applied = true
			return false
		}

		return true
	})

	if !applied {
		return fmt.Errorf("could not find statement at index %d, offset %d", spec.TargetIndex, spec.StartOffset)
	}
	return nil
}

func applyReplaceExpr(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec, value string) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		switch spec.NodeType {
		case "BinaryExpr.X":
			be, ok := n.(*ast.BinaryExpr)
			if !ok {
				return true
			}
			xPos := fset.Position(be.X.Pos())
			if xPos.Offset == spec.StartOffset {
				be.X = ast.NewIdent(value)
				applied = true
				return false
			}
		case "BinaryExpr.Y":
			be, ok := n.(*ast.BinaryExpr)
			if !ok {
				return true
			}
			yPos := fset.Position(be.Y.Pos())
			if yPos.Offset == spec.StartOffset {
				be.Y = ast.NewIdent(value)
				applied = true
				return false
			}
		case "IfStmt.Cond":
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok || ifStmt.Cond == nil {
				return true
			}
			condPos := fset.Position(ifStmt.Cond.Pos())
			if condPos.Offset == spec.StartOffset {
				ifStmt.Cond = ast.NewIdent(value)
				applied = true
				return false
			}
		case "ForStmt.Cond":
			forStmt, ok := n.(*ast.ForStmt)
			if !ok || forStmt.Cond == nil {
				return true
			}
			condPos := fset.Position(forStmt.Cond.Pos())
			if condPos.Offset == spec.StartOffset {
				forStmt.Cond = ast.NewIdent(value)
				applied = true
				return false
			}
		}

		return true
	})

	if !applied {
		return fmt.Errorf("could not find expression at offset %d for %s", spec.StartOffset, spec.NodeType)
	}
	return nil
}

func applySwapIfElse(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		pos := fset.Position(ifStmt.Body.Lbrace)
		if pos.Offset != spec.StartOffset {
			return true
		}

		elseBlock, ok := ifStmt.Else.(*ast.BlockStmt)
		if !ok {
			return true
		}

		// Swap the statement lists.
		ifStmt.Body.List, elseBlock.List = elseBlock.List, ifStmt.Body.List
		applied = true
		return false
	})

	if !applied {
		return fmt.Errorf("could not find if/else at offset %d for swap", spec.StartOffset)
	}
	return nil
}

func applySwapCase(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		var body *ast.BlockStmt
		switch sw := n.(type) {
		case *ast.SwitchStmt:
			body = sw.Body
		case *ast.TypeSwitchStmt:
			body = sw.Body
		default:
			return true
		}

		if body == nil {
			return true
		}

		idx1, idx2 := spec.TargetIndex, spec.TargetIndex2
		if idx1 < 0 || idx2 < 0 || idx1 >= len(body.List) || idx2 >= len(body.List) {
			return true
		}

		cc1, ok1 := body.List[idx1].(*ast.CaseClause)
		cc2, ok2 := body.List[idx2].(*ast.CaseClause)
		if !ok1 || !ok2 {
			return true
		}

		colonPos := fset.Position(cc1.Colon)
		if colonPos.Offset != spec.StartOffset {
			return true
		}

		// Swap the bodies.
		cc1.Body, cc2.Body = cc2.Body, cc1.Body
		applied = true
		return false
	})

	if !applied {
		return fmt.Errorf("could not find switch cases at indices [%d,%d], offset %d for swap",
			spec.TargetIndex, spec.TargetIndex2, spec.StartOffset)
	}
	return nil
}

func applyReturnReplace(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		pos := fset.Position(ret.Pos())
		if pos.Offset != spec.StartOffset {
			return true
		}

		if len(spec.ReturnMeta) != len(ret.Results) {
			return true
		}

		for i, repl := range spec.ReturnMeta {
			if repl == "" {
				continue
			}
			ret.Results[i] = ast.NewIdent(repl)
		}

		applied = true
		return false
	})

	if !applied {
		return fmt.Errorf("could not find return statement at offset %d", spec.StartOffset)
	}
	return nil
}

func applyReturnNegate(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		pos := fset.Position(ret.Pos())
		if pos.Offset != spec.StartOffset {
			return true
		}

		if spec.TargetIndex < 0 || spec.TargetIndex >= len(ret.Results) {
			return true
		}

		original := ret.Results[spec.TargetIndex]
		ret.Results[spec.TargetIndex] = &ast.UnaryExpr{
			Op: token.NOT,
			X:  original,
		}

		applied = true
		return false
	})

	if !applied {
		return fmt.Errorf("could not find return statement at offset %d for bool negation", spec.StartOffset)
	}
	return nil
}

func applyReplaceStmtWithReturn(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	var applied bool

	// Build the return statement.
	retStmt := &ast.ReturnStmt{}
	for _, val := range spec.ReturnMeta {
		if val != "" {
			retStmt.Results = append(retStmt.Results, ast.NewIdent(val))
		}
	}

	// Walk all block statements to find the target statement by offset and replace it.
	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		var stmts []ast.Stmt
		var setStmt func(int, ast.Stmt)

		switch node := n.(type) {
		case *ast.BlockStmt:
			stmts = node.List
			setStmt = func(i int, s ast.Stmt) { node.List[i] = s }
		case *ast.CaseClause:
			stmts = node.Body
			setStmt = func(i int, s ast.Stmt) { node.Body[i] = s }
		default:
			return true
		}

		for i, stmt := range stmts {
			pos := fset.Position(stmt.Pos())
			if pos.Offset == spec.StartOffset {
				setStmt(i, retStmt)
				applied = true
				return false
			}
		}

		return true
	})

	if !applied {
		return fmt.Errorf("could not find statement at offset %d for replacement with return", spec.StartOffset)
	}
	return nil
}

func applySwapCallArgs(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		lparenPos := fset.Position(call.Lparen)
		if lparenPos.Offset != spec.StartOffset {
			return true
		}

		idx1, idx2 := spec.TargetIndex, spec.TargetIndex2
		if idx1 < 0 || idx2 < 0 || idx1 >= len(call.Args) || idx2 >= len(call.Args) {
			return true
		}

		call.Args[idx1], call.Args[idx2] = call.Args[idx2], call.Args[idx1]
		applied = true
		return false
	})

	if !applied {
		return fmt.Errorf("could not find call expression at offset %d for argument swap", spec.StartOffset)
	}
	return nil
}

func applyOffsetExpr(fset *token.FileSet, file *ast.File, spec *mutation.StructuralSpec) error {
	offset := spec.TargetIndex // +1 or -1
	var applied bool

	ast.Inspect(file, func(n ast.Node) bool {
		if applied || n == nil {
			return false
		}

		switch spec.NodeType {
		case "IndexExpr.Index":
			ie, ok := n.(*ast.IndexExpr)
			if !ok {
				return true
			}
			pos := fset.Position(ie.Index.Pos())
			if pos.Offset != spec.StartOffset {
				return true
			}
			ie.Index = wrapWithOffset(ie.Index, offset)
			applied = true
			return false

		case "SliceExpr.High":
			se, ok := n.(*ast.SliceExpr)
			if !ok || se.High == nil {
				return true
			}
			pos := fset.Position(se.High.Pos())
			if pos.Offset != spec.StartOffset {
				return true
			}
			se.High = wrapWithOffset(se.High, offset)
			applied = true
			return false

		case "SliceExpr.Low":
			se, ok := n.(*ast.SliceExpr)
			if !ok || se.Low == nil {
				return true
			}
			pos := fset.Position(se.Low.Pos())
			if pos.Offset != spec.StartOffset {
				return true
			}
			se.Low = wrapWithOffset(se.Low, offset)
			applied = true
			return false
		}

		return true
	})

	if !applied {
		return fmt.Errorf("could not find expression at offset %d for %s offset", spec.StartOffset, spec.NodeType)
	}
	return nil
}

func wrapWithOffset(expr ast.Expr, offset int) ast.Expr {
	op := token.ADD
	if offset < 0 {
		op = token.SUB
	}
	return &ast.BinaryExpr{
		X:  expr,
		Op: op,
		Y:  &ast.BasicLit{Kind: token.INT, Value: "1"},
	}
}

// typeCheck performs a basic type check for noop generation.
// Uses a minimal configuration since we only need variable type info.
func typeCheck(fset *token.FileSet, file *ast.File) (*types.Package, *types.Info) {
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{
		// Ignore errors since we're working on potentially incomplete code.
		Error: func(err error) {},
	}
	pkg, _ := conf.Check("", fset, []*ast.File{file}, info)
	return pkg, info
}
