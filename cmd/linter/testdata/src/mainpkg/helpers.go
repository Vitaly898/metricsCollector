package main

import (
	"log"
	"os"
)

func init() {
	log.Fatal("init in main package") // want "log.Fatal is only allowed in main.main"
	os.Exit(1)                        // want "os.Exit is only allowed in main.main"
}
