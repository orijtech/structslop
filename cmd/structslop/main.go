package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/orijtech/structslop"
)

func main() {
	singlechecker.Main(structslop.Analyzer)
}
