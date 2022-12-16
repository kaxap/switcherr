package switcherr

import (
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"strings"
)

var errorType *types.Interface

func init() {
	errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
}

// Analyzer is the switcherr analysis.Analyzer instance.
var Analyzer = &analysis.Analyzer{
	Name: "switcherr",
	Doc:  "detect failed error handling with a switch statement",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if isGenerated(file) {
			//log.Printf("skipping generated file %s", pass.Fset.Position(file.Pos()).Filename)
			continue
		}

		// ast.Print(pass.Fset, file)
		v := &Visitor{
			fset:      pass.Fset,
			typesInfo: pass.TypesInfo,
		}
		ast.Walk(v, file)

		for _, e := range v.errors {
			pass.Reportf(e.Pos, "%s", e.ErrType)
		}
	}

	return nil, nil
}

func isGenerated(file *ast.File) bool {
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "// Code generated ") && strings.HasSuffix(c.Text, " DO NOT EDIT.") {
				return true
			}
		}
	}
	return false
}
