package aliases

import (
	logging "log"
	operating "os"
)

func helper() {
	logging.Fatal("aliased")        // want "log.Fatal is only allowed in main.main"
	logging.Fatalf("%s", "aliased") // want "log.Fatalf is only allowed in main.main"
	logging.Fatalln("aliased")      // want "log.Fatalln is only allowed in main.main"
	operating.Exit(1)               // want "os.Exit is only allowed in main.main"
}
