package termination

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

var initialized = func() bool {
	log.Fatal("package initializer") // want "log.Fatal is only allowed in main.main"
	os.Exit(1)
	return true
}()

func init() {
	log.Fatal("init") // want "log.Fatal is only allowed in main.main"
	os.Exit(1)
}

func helper() {
	log.Fatal("helper")        // want "log.Fatal is only allowed in main.main"
	log.Fatalf("%s", "helper") // want "log.Fatalf is only allowed in main.main"
	log.Fatalln("helper")      // want "log.Fatalln is only allowed in main.main"
	os.Exit(1)
	(log.Fatal)("paren") // want "log.Fatal is only allowed in main.main"
	(os.Exit)(1)
	defer log.Fatal("defer") // want "log.Fatal is only allowed in main.main"
	go os.Exit(1)
}

// A function named main in an ordinary package gets no exception.
func main() {
	log.Fatal("ordinary package") // want "log.Fatal is only allowed in main.main"
	os.Exit(1)
}

func allowed(t *testing.T) {
	log.Print("message")
	fmt.Print("message")
	_ = os.Getpid()
	_ = int(1)
	_ = len(make([]int, 1))
	t.Fatal("test failure")
	t.Fatalf("%s", "test failure")
	// The rule covers package functions, not methods of log.Logger.
	log.New(io.Discard, "", 0).Fatal("logger method")
}
