package structslop_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/orijtech/structslop"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, structslop.Analyzer, "struct")
}
