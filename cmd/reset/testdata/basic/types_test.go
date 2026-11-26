package sample

import "testing"

// generate:reset
type TestOnly struct{ Value int }

func TestScalarReset(t *testing.T) {
	s := Scalar{
		Signed: 1, Small: 2, Medium: 3, Large: 4, Largest: 5,
		Unsigned: 6, Byte: 7, UnsignedMedium: 8, UnsignedLarge: 9, UnsignedLargest: 10,
		Address: 11, Float: 12.5, Double: 13.5, Complex: 1 + 2i, DoubleComplex: 3 + 4i,
		Rune: 'x', First: 14, Second: 15, Active: true, Text: "text", Count: 16,
		External: 17, Embedded: true,
	}
	s.Reset()
	if s != (Scalar{}) {
		t.Fatalf("scalar was not zeroed: %+v", s)
	}
	var nilScalar *Scalar
	nilScalar.Reset()
	grouped := Grouped{Value: "text"}
	grouped.Reset()
	if grouped.Value != "" {
		t.Fatalf("grouped structure was not reset: %+v", grouped)
	}
	(&Empty{}).Reset()
}

func TestContainerReset(t *testing.T) {
	s := Containers{
		Values:   make([]string, 2, 8),
		Numbers:  Numbers{1, 2, 3},
		Labels:   Labels{"key": "value"},
		External: map[string]int{"external": 1},
	}
	s.Values[0] = "keep storage"
	values, numbers := s.Values, s.Numbers
	labels, external := s.Labels, s.External
	s.Reset()
	if len(s.Values) != 0 || cap(s.Values) != 8 || &s.Values[:cap(s.Values)][0] != &values[0] {
		t.Fatal("slice length or backing storage changed incorrectly")
	}
	if len(s.Numbers) != 0 || cap(s.Numbers) != cap(numbers) || &s.Numbers[:cap(s.Numbers)][0] != &numbers[0] {
		t.Fatal("named slice lost its storage")
	}
	if len(s.Labels) != 0 || len(s.External) != 0 || len(labels) != 0 || len(external) != 0 {
		t.Fatal("maps were not cleared in place")
	}
	s.Labels["new"] = "value"
	s.External["new"] = 2
	if labels["new"] != "value" || external["new"] != 2 {
		t.Fatal("map identity was not preserved")
	}
	if s.NilSlice != nil || s.NilMap != nil {
		t.Fatal("nil containers were allocated")
	}
	s.Reset()
	if len(s.Labels) != 0 || len(s.External) != 0 {
		t.Fatal("repeated reset did not clear maps")
	}
}
