package dotimports

import (
	. "log"
	. "os"
)

func helper() {
	Fatal("dot import")        // want "log.Fatal is only allowed in main.main"
	Fatalf("%s", "dot import") // want "log.Fatalf is only allowed in main.main"
	Fatalln("dot import")      // want "log.Fatalln is only allowed in main.main"
	Exit(1)                    // want "os.Exit is only allowed in main.main"
}
