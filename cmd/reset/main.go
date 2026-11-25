// Command reset generates Reset methods for structures marked with generate:reset.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	exitWithStatus(run(os.Args[1:], os.Stderr))
}

// run returns before process termination, so deferred cleanup can complete.
func run(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("reset", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dir := flags.String("dir", ".", "root directory to scan for Go packages")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: reset [-dir directory]")
		return 2
	}
	if err := generate(*dir); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func exitWithStatus(code int) {
	if code != 0 {
		os.Exit(code)
	}
}
