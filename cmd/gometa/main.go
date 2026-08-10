// Command gometa runs gometa's static declaration binding checks.
package main

import (
	"github.com/yvvlee/gometa/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
