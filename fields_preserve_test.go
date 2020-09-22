package structslop

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestStructFieldsPreserve(t *testing.T) {
	src := `package p
type s struct {
	_  [0]func()
	a3 [3]bool
	i int
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "struct_fields_preserve.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	if _, err := conf.Check("", fset, []*ast.File{f}, info); err != nil {
		t.Fatal(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if atyp, ok := n.(*ast.StructType); ok {
			if tv, ok := info.Types[atyp]; ok {
				styp := tv.Type.(*types.Struct)
				optStruct := optimalStructArrangement(styp)
				if optStruct.Field(0) != styp.Field(0) {
					t.Errorf("%v field order changed", styp.Field(0))
				}
			}
		}

		return true
	})
}
