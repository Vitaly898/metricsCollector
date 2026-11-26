package main

import (
	"flag"
	"fmt"
	"os"
)

// This is the original reset command, before its entry point was fixed.
func main() {
	dir := flag.String("dir", ".", "root directory to scan for Go packages")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: reset [-dir directory]")
		os.Exit(2) // want "os.Exit is forbidden in main.main"
	}
	if err := generate(*dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1) // want "os.Exit is forbidden in main.main"
	}
}

func generate(string) error { return nil }
