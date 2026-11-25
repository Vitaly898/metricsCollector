// Command linter runs the project's process termination checks.
package main

import (
	"github.com/Vitaly898/metricsCollector/internal/analyzer/exitcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(exitcheck.Analyzer)
}
