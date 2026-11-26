package sample

import (
	"bytes"
	clock "time"
	"unsafe"

	"example.com/sample/child"
)

// These names also exercise generated identifiers and import aliases.
type rs int
type zero int
type resetIndex int
type time int
type IntPointer *int
type RecursivePointer *RecursivePointer
type ChildAlias = child.Child

type Bare struct {
	Count int
	Names []string
}

type Manual struct {
	Value int
	Calls int
	Saved []int
}

func (m *Manual) Reset() {
	m.Value = 41
	m.Calls++
	m.Saved = m.Saved[:0]
}

type inherited struct{ Manual }

type WrongReset struct{ Value int }

func (*WrongReset) Reset(int) {}

type ValueReset struct{ Counter *int }

func (v ValueReset) Reset() {
	if v.Counter != nil {
		(*v.Counter)++
	}
}

// generate:reset
type Record struct {
	Count          *int
	Label          *string
	Flag           *bool
	Float          *float64
	Complex        *complex128
	Pointer        IntPointer
	Double         **int
	Recursive      RecursivePointer
	Slice          *[]int
	Map            *map[string]int
	Array          *[2]int
	NilPointer     *int
	NilSlice       *[]int
	NilMap         *map[string]int
	Child          *child.Child
	ChildValue     child.Child
	AliasedChild   *ChildAlias
	Self           *Record
	Manual         Manual
	ManualPointer  *Manual
	ExternalManual child.Manual
	Inherited      inherited
	ValueMethod    ValueReset
	WrongMethod    WrongReset
	Bare           Bare
	BarePointer    *Bare
	Anonymous      struct {
		Flag  bool
		Items []string
	}
	Foreign        clock.Time
	ForeignPointer *clock.Time
	Buffer         bytes.Buffer
	ArraySlices    [2][]int
	ArrayMaps      [2]map[string]int
	ArrayChildren  [2]child.Child
	ArrayPointers  [2]*int
	Matrix         [2][2]int
	Any            any
	Handler        func() int
	Channel        chan int
	Interface      interface{ Reset() }
	Unsafe         unsafe.Pointer
}

// generate:reset
type MarkedManual struct{ Calls int }

func (m *MarkedManual) Reset() { m.Calls++ }

// generate:reset
type Box[rs any] struct {
	Value   rs
	Pointer *rs
	Items   []rs
	Child   *Box[rs]
}

// generate:reset
type Pair[A, B any] struct {
	First  A
	Second B
}

// These calls must compile during the first generator run, too.
func use(record *Record) {
	record.Reset()
	record.Child.Reset()
}
