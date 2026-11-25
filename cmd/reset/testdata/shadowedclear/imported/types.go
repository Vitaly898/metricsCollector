package imported

import other "example.com/sample/clear"

// generate:reset
type Container struct {
	Entries map[string]int
	Plain   other.Plain
}
