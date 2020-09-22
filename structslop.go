// Copyright 2020 Orijtech, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package structslop

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"sort"

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

		r := checkSloppy(styp)
		if !r.sloppy() {
			return
		}

		pass.Report(analysis.Diagnostic{
			Pos:     n.Pos(),
			End:     n.End(),
			Message: fmt.Sprintf("%v has size %d, could be %d, rearrange to %v for optimal size", styp, r.oldSize, r.newSize, r.suggestedStruct),
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
	})
	return nil, nil
}

var sizes = types.SizesFor(build.Default.Compiler, build.Default.GOARCH)

type result struct {
	oldSize         int64
	newSize         int64
	suggestedStruct *types.Struct
}

func (r result) sloppy() bool {
	return r.oldSize > r.newSize
}

func checkSloppy(origStruct *types.Struct) result {
	optStruct := optimalStructArrangement(origStruct)
	return result{
		oldSize:         sizes.Sizeof(origStruct),
		newSize:         sizes.Sizeof(optStruct),
		suggestedStruct: optStruct,
	}
}

func optimalStructArrangement(s *types.Struct) *types.Struct {
	nf := s.NumFields()
	fields := make([]*types.Var, nf)
	for i := 0; i < nf; i++ {
		fields[i] = s.Field(i)
	}

	sort.Slice(fields, func(i, j int) bool {
		ti, tj := fields[i].Type(), fields[j].Type()
		si, sj := sizes.Sizeof(ti), sizes.Sizeof(tj)

		if si == 0 && sj != 0 {
			return true
		}
		if sj == 0 && si != 0 {
			return false
		}

		ai, aj := sizes.Alignof(ti), sizes.Alignof(tj)
		if ai != aj {
			return ai > aj
		}

		if si != sj {
			return si > sj
		}

		return false
	})

	return types.NewStruct(fields, nil)
}
