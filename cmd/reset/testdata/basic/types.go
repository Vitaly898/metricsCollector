package sample

import "example.com/sample/child"

type Count int64
type Text = string
type Numbers []int
type Labels map[string]string
type Embedded bool

// Scalar includes all primitive numeric categories and shared field declarations.
// generate:reset
type Scalar struct {
	Signed          int
	Small           int8
	Medium          int16
	Large           int32
	Largest         int64
	Unsigned        uint
	Byte            byte
	UnsignedMedium  uint16
	UnsignedLarge   uint32
	UnsignedLargest uint64
	Address         uintptr
	Float           float32
	Double          float64
	Complex         complex64
	DoubleComplex   complex128
	Rune            rune
	First, Second   int
	Active          bool
	Text            Text
	Count           Count
	External        child.Number
	Embedded
	_ [4]byte
}

// generate:reset
type Containers struct {
	Values   []string
	Numbers  Numbers
	Labels   Labels
	External child.Labels
	NilSlice []int
	NilMap   map[string]int
}

type Unmarked struct {
	Value int
}

// The directive belongs to this function, not the type declared inside it.
// generate:reset
func helper() {
	type Local struct{ Value int }
}
