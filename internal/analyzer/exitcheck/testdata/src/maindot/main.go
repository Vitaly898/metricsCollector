package main

import . "os"

func main() {
	Exit(1) // want "os.Exit is forbidden in main.main"
}
