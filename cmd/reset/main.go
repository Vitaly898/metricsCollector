// Command reset generates Reset methods for structures marked with generate:reset.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	dir := flag.String("dir", ".", "root directory to scan for Go packages")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: reset [-dir directory]")
		os.Exit(2)
	}
	if err := generate(*dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
