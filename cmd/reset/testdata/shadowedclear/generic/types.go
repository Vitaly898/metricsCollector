package generic

// generate:reset
type Container[clear any] struct {
	Value   clear
	Entries map[string]int
}
