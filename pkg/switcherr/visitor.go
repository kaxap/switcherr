package switcherr

import (
	"go/ast"
	"go/token"
	"go/types"
)

type Visitor struct {
	fset      *token.FileSet
	typesInfo *types.Info
	errors    []ErrorItem
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	sw, ok := n.(*ast.SwitchStmt)
	if !ok {
		return v
	}

	caseIndexes := map[CaseType]int{}
	updateIndex := func(typ CaseType, index int) {
		if typ == CaseErrorIs {
			// always overwrite CaseErrorIs since, error.Is must be higher than err != nil
			caseIndexes[typ] = index
			return
		}
		if _, ok := caseIndexes[typ]; ok {
			return
		}
		caseIndexes[typ] = index
	}

	for i, c := range sw.Body.List {
		caseBody := c.(*ast.CaseClause).List
		if caseBody == nil {
			// default case
			if v.hasAnyErrorIndent(c) {
				// found a default case that handles an error
				updateIndex(CaseErrNeqNil, i)
			}
			continue
		}
		for _, e := range caseBody {
			switch a := e.(type) {
			case *ast.BinaryExpr:
				switch {
				case binaryExprLookup(a, v.isErrNilCheck):
					updateIndex(CaseErrorEqNil, i)
				case binaryExprLookup(a, v.isErrNonNilCheck):
					updateIndex(CaseErrNeqNil, i)
				case binaryExprLookup(a, func(a *ast.BinaryExpr) bool { return v.isErrorPkgFun(a.X) || v.isErrorPkgFun(a.Y) }):
					updateIndex(CaseErrorIs, i)
				default:
					updateIndex(CaseNotErrorHandler, i)
				}
			case *ast.CallExpr:
				if v.isErrorPkgFun(a) {
					updateIndex(CaseErrorIs, i)
				} else {
					updateIndex(CaseNotErrorHandler, i)
				}
			default:
				updateIndex(CaseNotErrorHandler, i)
			}
		}
	}
	errTyp := v.getError(caseIndexes)
	if errTyp != ErrorTypeNoError {
		v.errors = append(v.errors, ErrorItem{
			Pos:     sw.Pos(),
			ErrType: errTyp,
		})
	}
	return v
}

// hasAnyErrorIndent recursively searches an error typed indent in the block if any.
func (v *Visitor) hasAnyErrorIndent(c ast.Stmt) bool {
	if c == nil {
		return false
	}
	switch a := c.(type) {
	case *ast.BlockStmt:
		for _, s := range a.List {
			if v.hasAnyErrorIndent(s) {
				return true
			}
		}
	case *ast.IfStmt:
		return v.hasAnyErrorIndent(a.Body) || v.hasAnyErrorIndent(a.Else)
	case *ast.SwitchStmt:
		return v.hasAnyErrorIndent(a.Body)
	case *ast.CaseClause:
		for _, s := range a.Body {
			if v.hasAnyErrorIndent(s) {
				return true
			}
		}
	case *ast.AssignStmt:
		for _, e := range a.Rhs {
			if v.isError(e) {
				return true
			}

		}
	case *ast.ExprStmt:
		return v.isError(a.X)
	case *ast.ReturnStmt:
		for _, e := range a.Results {
			if v.isError(e) {
				return true
			}
		}
	}
	return false
}

func (v *Visitor) getError(m map[CaseType]int) ErrorType {
	indErrorIs, existErrorIs := m[CaseErrorIs]
	indErrNeqNil, existErrNeqNil := m[CaseErrNeqNil]
	indNotErrorHandler, existNotErrorHandler := m[CaseNotErrorHandler]

	if !existErrorIs && !existErrNeqNil {
		// not an error-checking switch
		return ErrorTypeNoError
	}
	if !existErrNeqNil {
		// no err != nil general case handling
		return ErrorTypeNoNeqNil
	}
	if !existErrorIs {
		// we there are no error type checks then err != nil should be the first case
		if indErrNeqNil != 0 {
			return ErrorTypeNeqNilAfterNonError
		}
		return ErrorTypeNoError
	}
	if indErrorIs > indErrNeqNil {
		// error.Is after err != nil
		return ErrorTypeIsAfterNeqNil
	}
	if existNotErrorHandler && indErrNeqNil > indNotErrorHandler {
		return ErrorTypeNeqNilAfterNonError
	}
	return ErrorTypeNoError
}

func (v *Visitor) getInfo(a ast.Node) string {
	if v.fset != nil {
		return v.fset.Position(a.Pos()).String()
	} else {
		return "no fset info"
	}
}

func (v *Visitor) getType(e ast.Expr) types.Type {
	return v.typesInfo.Types[e].Type
}

func (v *Visitor) isErrorType(t types.Type) bool {
	return types.Implements(t, errorType)
}

func (v *Visitor) isError(e ast.Expr) bool {
	if v.typesInfo == nil {
		return false
	}
	if typ, ok := v.typesInfo.Types[e]; ok {
		return v.isErrorType(typ.Type)
	}
	return false
}

func (v *Visitor) isNilIdent(n ast.Node) bool {
	a, ok := n.(*ast.Ident)
	if !ok {
		return false
	}
	return a.Name == "nil"
}

func (v *Visitor) isErrorPkgFun(n ast.Node) bool {
	a, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}

	fun, ok := a.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := fun.X.(*ast.Ident)
	pkgName := ""
	if ok {
		pkgName = pkgIdent.Name
	}
	funName := fun.Sel.Name
	if pkgName == "errors" && (funName == "Is" || funName == "As") {
		return true
	}

	for _, arg := range a.Args {
		if argIdent, ok := arg.(*ast.Ident); ok {
			if v.isError(argIdent) {
				return true
			}
		}

	}

	return false
}

func (v *Visitor) isErrNonNilCheck(n *ast.BinaryExpr) bool {
	// looking for "!="
	if n.Op != token.NEQ {
		return false
	}

	// either err != nil or nil != err
	if v.isError(n.X) && v.isNilIdent(n.Y) {
		return true
	}
	if v.isNilIdent(n.X) && v.isError(n.Y) {
		return true
	}
	return false
}

func (v *Visitor) isErrNilCheck(n *ast.BinaryExpr) bool {
	// looking for "=="
	if n.Op != token.EQL {
		return false
	}

	// either err == nil or nil == err
	if v.isError(n.X) && v.isNilIdent(n.Y) {
		return true
	}
	if v.isNilIdent(n.X) && v.isError(n.Y) {
		return true
	}
	return false
}

// binaryExprLookup lookups binary expression tree if any needle(node) is true
func binaryExprLookup(stack *ast.BinaryExpr, testFunc func(a *ast.BinaryExpr) bool) bool {

	if x, ok := stack.X.(*ast.BinaryExpr); ok {
		// this is a tree of binary expressions, such as a || b || c
		// this will produce a binary expression tree root.l = a, root.r = node, none.l = b, node.r = c
		if binaryExprLookup(x, testFunc) {
			// one of the nodes down the line is satisfies condition of testFunc
			return true
		}
		if testFunc(x) {
			return true
		}
	}

	// check the same for a.Y
	if y, ok := stack.Y.(*ast.BinaryExpr); ok {
		if binaryExprLookup(y, testFunc) {
			return true
		}
		if testFunc(y) {
			return true
		}
	}
	return testFunc(stack)
}
