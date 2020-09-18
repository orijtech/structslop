package structslop

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for structs that can be rearrange fields to provide for maximum space/allocation efficiency`

// Analyzer describes struct slop analysis function detector.
var Analyzer = &analysis.Analyzer{
	Name:     "structslop",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		s := n.(*ast.StructType)
		if s.Fields.NumFields() < 2 {
			return
		}
		pass.Reportf(s.Pos(), "not implemented")
	})
	return nil, nil
}
