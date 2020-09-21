package structslop

import (
	"fmt"
	"go/ast"
	"go/types"

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
		atyp := n.(*ast.StructType)
		styp, ok := pass.TypesInfo.Types[atyp].Type.(*types.Struct)
		// Type information may be incomplete.
		if !ok {
			return
		}

		if styp.NumFields() < 2 {
			return
		}

		if r := malign(styp); r.Slop() {
			pass.Report(analysis.Diagnostic{
				Pos:     n.Pos(),
				End:     n.End(),
				Message: fmt.Sprintf("%v has size %d, could be %d", styp, r.oldSize, r.newSize),
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: fmt.Sprintf("Rearrange struct fields: %v", r.suggestedStruct),
					TextEdits: []analysis.TextEdit{
						{
							Pos:     n.Pos(),
							End:     n.End(),
							NewText: []byte(fmt.Sprintf("%v", r.suggestedStruct)),
						},
					},
				}},
			})
		}
	})
	return nil, nil
}

type result struct {
	oldSize         int64
	newSize         int64
	suggestedStruct *types.Struct
}

func (r result) Slop() bool {
	return r.oldSize > r.newSize
}
