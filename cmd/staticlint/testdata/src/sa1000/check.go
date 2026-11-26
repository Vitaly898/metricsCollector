package sa1000

import "regexp"

func bad() {
	_, _ = regexp.Compile("[") // want "error parsing regexp"
}
