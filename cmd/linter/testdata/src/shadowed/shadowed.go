package shadowed

import (
	"log"
	"os"
)

func panic(any) {}

type logger struct{}

func (logger) Fatal(...any)          {}
func (logger) Fatalf(string, ...any) {}
func (logger) Fatalln(...any)        {}

type process struct{}

func (process) Exit(int) {}

func helper() {
	panic("custom function")
	log.Print("ordinary logging")
	_ = os.Getpid()

	log := logger{}
	os := process{}
	log.Fatal("shadowed import")
	log.Fatalf("%s", "shadowed import")
	log.Fatalln("shadowed import")
	os.Exit(1)

	func(panic func(any)) {
		panic("parameter")
	}(func(any) {})
}

func localPanic() {
	panic := func(any) {}
	panic("local function")
}
