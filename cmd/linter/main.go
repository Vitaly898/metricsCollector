// Command linter forbids builtin panic and restricts process termination to main.main.
package main

import "golang.org/x/tools/go/analysis/singlechecker"

func main() {
	singlechecker.Main(Analyzer)
}
