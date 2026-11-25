package shadowedclear

var customClearCalls int

// A valid user function that intentionally leaves the map unchanged.
func clear(map[string]int) { customClearCalls++ }

type Labels map[string]int

// generate:reset
type Container struct {
	Values  map[string]int
	Labels  Labels
	Pointer *map[string]int
	Array   [1]map[string]int
	NilMap  map[string]int
}
