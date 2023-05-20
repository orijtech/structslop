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
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	includeTestFiles bool
	verbose          bool
	apply            bool
	generated        bool
)

func init() {
	Analyzer.Flags.BoolVar(&includeTestFiles, "include-test-files", includeTestFiles, "also check test files")
	Analyzer.Flags.BoolVar(&verbose, "verbose", verbose, "print all information, even when struct is not sloppy")
	Analyzer.Flags.BoolVar(&apply, "apply", apply, "apply suggested fixes (using -fix won't work)")
	Analyzer.Flags.BoolVar(&generated, "generated", generated, "report issues in generated code")
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

	dec := decorator.NewDecorator(pass.Fset)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.StructType)(nil),
	}

	fileDiags := make(map[string][]byte)
	var af *ast.File
	var df *dst.File

	// Track generated files unless -generated is set.
	genFiles := make(map[*token.File]bool)
	if !generated {
	files:
		for _, f := range pass.Files {
			for _, c := range f.Comments {
				for _, l := range c.List {
					if strings.HasPrefix(l.Text, "// Code generated ") && strings.HasSuffix(l.Text, " DO NOT EDIT.") {
						file := pass.Fset.File(f.Pos())
						genFiles[file] = true
						continue files
					}
				}
			}
		}
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		file := pass.Fset.File(n.Pos())
		if strings.HasSuffix(file.Name(), "_test.go") && !includeTestFiles {
			return
		}
		// Skip generated structs if instructed.
		if !generated && genFiles[file] {
			return
		}
		if f, ok := n.(*ast.File); ok {
			af = f
			df, _ = dec.DecorateFile(af)
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

		var buf bytes.Buffer
		expr, err := parser.ParseExpr(formatStruct(r.optStruct, pass.Pkg.Path()))
		if err != nil {
			return
		}
		if err := format.Node(&buf, token.NewFileSet(), expr.(*ast.StructType)); err != nil {
			return
		}

		var msg string
		switch {
		case r.oldGcSize == r.newGcSize:
			msg = fmt.Sprintf("struct has size %d (size class %d)", r.oldGcSize, r.oldRuntimeSize)
		case r.oldGcSize != r.newGcSize:
			msg = fmt.Sprintf(
				"struct has size %d (size class %d), could be %d (size class %d), optimal fields order:\n%s\n",
				r.oldGcSize,
				r.oldRuntimeSize,
				r.newGcSize,
				r.newRuntimeSize,
				buf.String(),
			)
			if r.sloppy() {
				msg = fmt.Sprintf(
					"struct has size %d (size class %d), could be %d (size class %d), you'll save %.2f%% if you rearrange it to:\n%s\n",
					r.oldGcSize,
					r.oldRuntimeSize,
					r.newGcSize,
					r.newRuntimeSize,
					r.savings(),
					buf.String(),
				)
			}
		}

		dtyp := dec.Dst.Nodes[atyp].(*dst.StructType)
		fields := make([]*dst.Field, 0, len(r.optIdx))
		dummy := &dst.Field{}
		for _, f := range dtyp.Fields.List {
			fields = append(fields, f)
			if len(f.Names) == 0 {
				continue
			}
			for range f.Names[1:] {
				fields = append(fields, dummy)
			}
		}
		optFields := make([]*dst.Field, 0, len(r.optIdx))
		for _, i := range r.optIdx {
			f := fields[i]
			if f == dummy {
				continue
			}
			optFields = append(optFields, f)
		}
		dtyp.Fields.List = optFields

		var suggested bytes.Buffer
		if err := decorator.Fprint(&suggested, df); err != nil {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos:            n.Pos(),
			End:            n.End(),
			Message:        msg,
			SuggestedFixes: nil,
		})
		f := pass.Fset.File(n.Pos())
		fileDiags[f.Name()] = suggested.Bytes()
	})

	if !apply {
		return nil, nil
	}
	for fn, content := range fileDiags {
		fi, err := os.Open(fn)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to open file: %v", err)
		}
		st, err := fi.Stat()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to get file stat: %v", err)
		}
		if err := fi.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close file: %v", err)
		}
		if err := os.WriteFile(fn, content, st.Mode()); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to write suggested fix to file: %v", err)
		}
	}
	return nil, nil
}

type result struct {
	oldGcSize      int64
	newGcSize      int64
	oldRuntimeSize int64
	newRuntimeSize int64
	optStruct      *types.Struct
	optIdx         []int
}

func (r result) sloppy() bool {
	return r.oldRuntimeSize > r.newRuntimeSize
}

func (r result) savings() float64 {
	return float64(r.oldRuntimeSize-r.newRuntimeSize) / float64(r.oldRuntimeSize) * 100
}

func mapFieldIdx(s *types.Struct) map[*types.Var]int {
	m := make(map[*types.Var]int, s.NumFields())
	for i := 0; i < s.NumFields(); i++ {
		m[s.Field(i)] = i
	}
	return m
}

func checkSloppy(pass *analysis.Pass, origStruct *types.Struct) result {
	m := mapFieldIdx(origStruct)
	optStruct := optimalStructArrangement(pass.TypesSizes, m)
	idx := make([]int, optStruct.NumFields())
	for i := range idx {
		idx[i] = m[optStruct.Field(i)]
	}
	r := result{
		oldGcSize: pass.TypesSizes.Sizeof(origStruct),
		newGcSize: pass.TypesSizes.Sizeof(optStruct),
		optStruct: optStruct,
		optIdx:    idx,
	}
	r.oldRuntimeSize = int64(roundUpSize(uintptr(r.oldGcSize)))
	r.newRuntimeSize = int64(roundUpSize(uintptr(r.newGcSize)))
	return r
}

func optimalStructArrangement(sizes types.Sizes, m map[*types.Var]int) *types.Struct {
	fields := make([]*types.Var, len(m))
	for v, i := range m {
		fields[i] = v
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
