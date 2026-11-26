package ineffassign

func bad() int {
	value := 1 // want "ineffectual assignment to value"
	value = 2
	return value
}
