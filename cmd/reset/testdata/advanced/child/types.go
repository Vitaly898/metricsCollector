package child

// generate:reset
type Child struct {
	Name   string
	Values []int
}

type Manual struct {
	Value int
	Calls int
}

func (m *Manual) Reset() {
	m.Value = 42
	m.Calls++
}

func use(c *Child) { c.Reset() }
