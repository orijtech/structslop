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
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	includeTestFiles bool
	verbose          bool
)

func init() {
	Analyzer.Flags.BoolVar(&includeTestFiles, "include-test-files", includeTestFiles, "also check test files")
	Analyzer.Flags.BoolVar(&verbose, "verbose", verbose, "print all information, even when struct is not sloppy")
}

const Doc = `check for structs that can be rearrange fields to provide for maximum space/allocation efficiency`

// Analyzer describes struct slop analysis function detector.
var Analyzer = &analysis.Analyzer{
	Name:     "structslop",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Use custom sizes instance, which implements types.Sizes for calculating struct size.
	// go/types and gc does not agree about the struct size.
	// See https://github.com/golang/go/issues/14909#issuecomment-199936232
	pass.TypesSizes = &sizes{
		stdSizes: types.SizesFor(build.Default.Compiler, build.Default.GOARCH),
		maxAlign: pass.TypesSizes.Alignof(types.Typ[types.UnsafePointer]),
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		if strings.HasSuffix(pass.Fset.File(n.Pos()).Name(), "_test.go") && !includeTestFiles {
			return
		}
		atyp := n.(*ast.StructType)
		styp, ok := pass.TypesInfo.Types[atyp].Type.(*types.Struct)
		// Type information may be incomplete.
		if !ok {
			return
		}
		if !verbose && styp.NumFields() < 2 {
			return
		}

		r := checkSloppy(pass, styp)
		if !verbose && !r.sloppy() {
			return
		}

		var msg string
		switch {
		case r.oldGcSize == r.newGcSize:
			msg = fmt.Sprintf("struct has size %d (size class %d)", r.oldGcSize, r.oldRuntimeSize)
		case r.oldGcSize != r.newGcSize:
			curPkgPath := pass.Pkg.Path()
			optStyp := formatStruct(r.suggestedStruct, curPkgPath)
			msg = fmt.Sprintf(
				"struct has size %d (size class %d), could be %d (size class %d), rearrange to %s for optimal size",
				r.oldGcSize,
				r.oldRuntimeSize,
				r.newGcSize,
				r.newRuntimeSize,
				optStyp,
			)
			if r.sloppy() {
				msg = fmt.Sprintf(
					"struct has size %d (size class %d), could be %d (size class %d), rearrange to %s for optimal size (%.2f%% savings)",
					r.oldGcSize,
					r.oldRuntimeSize,
					r.newGcSize,
					r.newRuntimeSize,
					optStyp,
					r.savings(),
				)
			}
		}
		pass.Report(analysis.Diagnostic{
			Pos:            n.Pos(),
			End:            n.End(),
			Message:        msg,
			SuggestedFixes: nil,
		})
	})
	return nil, nil
}

type result struct {
	oldGcSize       int64
	newGcSize       int64
	oldRuntimeSize  int64
	newRuntimeSize  int64
	suggestedStruct *types.Struct
}

func (r result) sloppy() bool {
	return r.oldRuntimeSize > r.newRuntimeSize
}

func (r result) savings() float64 {
	return float64(r.oldRuntimeSize-r.newRuntimeSize) / float64(r.oldRuntimeSize) * 100
}

func checkSloppy(pass *analysis.Pass, origStruct *types.Struct) result {
	optStruct := optimalStructArrangement(pass.TypesSizes, origStruct)
	r := result{
		oldGcSize:       pass.TypesSizes.Sizeof(origStruct),
		newGcSize:       pass.TypesSizes.Sizeof(optStruct),
		suggestedStruct: optStruct,
	}
	r.oldRuntimeSize = int64(roundUpSize(uintptr(r.oldGcSize)))
	r.newRuntimeSize = int64(roundUpSize(uintptr(r.newGcSize)))
	return r
}

func optimalStructArrangement(sizes types.Sizes, s *types.Struct) *types.Struct {
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

func formatStruct(styp *types.Struct, curPkgPath string) string {
	qualifier := func(p *types.Package) string {
		if p.Path() == curPkgPath {
			return ""
		}
		return p.Name()
	}
	return types.TypeString(styp, qualifier)
}
