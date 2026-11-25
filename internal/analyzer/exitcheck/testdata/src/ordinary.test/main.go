package main

import "os"

func main() {
	os.Exit(0) // want "os.Exit is forbidden in main.main"
}
