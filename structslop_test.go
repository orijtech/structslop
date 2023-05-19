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

package structslop_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/orijtech/structslop"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, structslop.Analyzer, "struct")
}

func TestApply(t *testing.T) {
	dir := strings.Join([]string{".", "testdata", "src"}, string(os.PathSeparator))
	tmpdir, err := os.MkdirTemp(dir, "structslop-test-apply-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	fn := filepath.Join(tmpdir, "p.go")
	src, _ := os.ReadFile(filepath.Join(".", "testdata", "src", "struct", "p.go"))
	if err := os.WriteFile(fn, src, 0644); err != nil {
		t.Fatal(err)
	}
	testdata := analysistest.TestData()
	_ = structslop.Analyzer.Flags.Set("apply", "true")
	defer func() {
		_ = structslop.Analyzer.Flags.Set("apply", "false")
	}()
	analysistest.Run(t, testdata, structslop.Analyzer, filepath.Base(tmpdir))
	got, _ := os.ReadFile(fn)
	expected, _ := os.ReadFile(filepath.Join(".", "testdata", "src", "struct", "p.go.golden"))
	if !bytes.Equal(expected, got) {
		t.Errorf("unexpected suggested fix, want:\n%s\ngot:\n%s\n", string(expected), string(got))
	}
}

func TestIncludeTestFiles(t *testing.T) {
	testdata := analysistest.TestData()
	_ = structslop.Analyzer.Flags.Set("include-test-files", "true")
	defer func() {
		_ = structslop.Analyzer.Flags.Set("include-test-files", "false")
	}()
	analysistest.Run(t, testdata, structslop.Analyzer, "include-test-files")
}

func TestVerboseMode(t *testing.T) {
	testdata := analysistest.TestData()
	_ = structslop.Analyzer.Flags.Set("verbose", "true")
	defer func() {
		_ = structslop.Analyzer.Flags.Set("verbose", "false")
	}()
	analysistest.Run(t, testdata, structslop.Analyzer, "verbose")
}

func TestGenerated(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, structslop.Analyzer, "generated")
}
