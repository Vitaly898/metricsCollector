package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed")
	log.Fatalf("%s", "allowed")
	log.Fatalln("allowed")
	os.Exit(1)
	(log.Fatal)("allowed")
	(os.Exit)(1)
	panic("main") // want "use of builtin panic is forbidden"

	callback := func() {
		log.Fatal("closure") // want "log.Fatal is only allowed in main.main"
		os.Exit(1)           // want "os.Exit is only allowed in main.main"
		panic("closure")     // want "use of builtin panic is forbidden"
		func() {
			os.Exit(1) // want "os.Exit is only allowed in main.main"
		}()
	}
	callback()
	defer func() {
		log.Fatalf("%s", "deferred closure") // want "log.Fatalf is only allowed in main.main"
	}()
	go func() {
		os.Exit(1) // want "os.Exit is only allowed in main.main"
	}()
	// Leaving a closure restores the context of main.
	log.Fatal("allowed after closure")
	os.Exit(1)
}

func helper() {
	log.Fatal("helper after main") // want "log.Fatal is only allowed in main.main"
	os.Exit(1)                     // want "os.Exit is only allowed in main.main"
}

type worker struct{}

// A method named main gets no exception, even in package main.
func (worker) main() {
	log.Fatal("method") // want "log.Fatal is only allowed in main.main"
	os.Exit(1)          // want "os.Exit is only allowed in main.main"
}
