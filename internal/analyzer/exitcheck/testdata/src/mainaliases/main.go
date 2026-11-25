package main

import (
	logging "log"
	operating "os"
)

func main() {
	logging.Fatal("allowed")
	operating.Exit(1) // want "os.Exit is forbidden in main.main"
	os := struct{ Exit func(int) }{Exit: func(int) {}}
	os.Exit(1)
}
