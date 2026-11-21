package sample

import (
	"testing"
	clock "time"
	"unsafe"

	"example.com/sample/child"
)

func TestPointers(t *testing.T) {
	count, label, flag := 12, "text", true
	float, complex := 12.5, complex128(1+2i)
	values, entries := []int{1, 2, 3}, map[string]int{"key": 1}
	array := [2]int{1, 2}
	countPointer := &count
	leafValue := 17
	r := Record{
		Count: &count, Label: &label, Flag: &flag, Float: &float, Complex: &complex,
		Pointer: &count, Double: &countPointer, Slice: &values, Map: &entries, Array: &array,
		Self: &Record{Count: &leafValue},
	}
	leaf := r.Self
	backing := &values[0]
	mapAlias := entries
	r.Reset()
	if count != 0 || label != "" || flag || float != 0 || complex != 0 || array != ([2]int{}) || leafValue != 0 {
		t.Fatal("values under pointers were not reset")
	}
	if r.Count != &count || r.Label != &label || r.Flag != &flag || r.Float != &float || r.Complex != &complex ||
		r.Pointer != &count || r.Double != &countPointer || *r.Double != &count || r.Slice != &values ||
		r.Map != &entries || r.Array != &array || r.Self != leaf {
		t.Fatal("pointer identity was not preserved")
	}
	if len(values) != 0 || &values[:cap(values)][0] != backing || len(entries) != 0 || len(mapAlias) != 0 {
		t.Fatal("containers under pointers lost their storage or contents were not cleared")
	}
	entries["new"] = 2
	if mapAlias["new"] != 2 {
		t.Fatal("map under pointer was replaced")
	}
	var nilValues []int
	var nilEntries map[string]int
	var end RecursivePointer
	middle := RecursivePointer(&end)
	recursive := RecursivePointer(&middle)
	nilLevel := new(*int)
	r = Record{Double: nilLevel, NilSlice: &nilValues, NilMap: &nilEntries, Recursive: recursive}
	r.Reset()
	if r.Double != nilLevel || *r.Double != nil || r.NilPointer != nil || r.NilSlice != &nilValues ||
		r.NilMap != &nilEntries || nilValues != nil || nilEntries != nil || r.Child != nil || r.Self != nil {
		t.Fatal("reset allocated a nil pointer or container")
	}
	if r.Recursive != recursive || *recursive != middle || *middle != nil {
		t.Fatal("recursive pointer types were not preserved")
	}
	var nilRecord *Record
	nilRecord.Reset()
}

func TestNestedResetMethods(t *testing.T) {
	manual := &Manual{Value: 7, Saved: []int{1, 2}}
	c := &child.Child{Name: "child", Values: []int{1, 2, 3}}
	alias := &ChildAlias{Name: "alias", Values: []int{4}}
	var calls int
	r := Record{
		Child: c, ChildValue: child.Child{Name: "value", Values: []int{8}}, AliasedChild: alias,
		Manual: Manual{Value: 9, Saved: []int{3}}, ManualPointer: manual,
		ExternalManual: child.Manual{Value: 5}, Inherited: inherited{Manual: Manual{Value: 8}},
		ValueMethod: ValueReset{Counter: &calls}, WrongMethod: WrongReset{Value: 1},
	}
	r.Reset()
	if r.Child != c || r.AliasedChild != alias || c.Name != "" || alias.Name != "" || r.ChildValue.Name != "" ||
		len(c.Values) != 0 || c.Values == nil || len(r.ChildValue.Values) != 0 || r.ChildValue.Values == nil {
		t.Fatal("generated nested Reset methods were not called")
	}
	if r.Manual.Value != 41 || r.Manual.Calls != 1 || len(r.Manual.Saved) != 0 || r.Manual.Saved == nil ||
		manual.Value != 41 || manual.Calls != 1 || r.ManualPointer != manual ||
		r.ExternalManual.Value != 42 || r.ExternalManual.Calls != 1 || r.Inherited.Value != 41 ||
		r.Inherited.Calls != 1 || calls != 1 || r.WrongMethod.Value != 0 {
		t.Fatal("handwritten, imported, promoted or value Reset method was not respected")
	}
	marked := &MarkedManual{}
	marked.Reset()
	if marked.Calls != 1 {
		t.Fatal("handwritten Reset on a marked type was replaced")
	}
}

func TestOtherTypes(t *testing.T) {
	stamp := clock.Now()
	bare := &Bare{Count: 8, Names: []string{"clear"}}
	number := 15
	r := Record{
		Bare: Bare{Count: 9, Names: []string{"clear"}}, BarePointer: bare,
		Foreign: stamp, ForeignPointer: &stamp,
		ArraySlices: [2][]int{{1, 2}, nil}, ArrayMaps: [2]map[string]int{{"key": 1}, nil},
		ArrayChildren: [2]child.Child{{Name: "array", Values: []int{1}}, {}},
		ArrayPointers: [2]*int{&number, nil}, Matrix: [2][2]int{{1, 2}, {3, 4}},
		Any: "value", Handler: func() int { return 1 }, Channel: make(chan int),
		Interface: (*Manual)(nil), Unsafe: unsafe.Pointer(&number),
	}
	r.Anonymous.Flag = true
	r.Anonymous.Items = []string{"clear"}
	r.Buffer.WriteString("buffer")
	bufferCap := r.Buffer.Cap()
	backing := &r.ArraySlices[0][0]
	mapAlias := r.ArrayMaps[0]
	r.Reset()
	if r.Bare.Count != 0 || r.Bare.Names != nil || r.BarePointer != bare || bare.Count != 0 || bare.Names != nil ||
		r.Anonymous.Flag || r.Anonymous.Items != nil || !r.Foreign.IsZero() || r.ForeignPointer != &stamp || !stamp.IsZero() {
		t.Fatal("structures without Reset were not zeroed")
	}
	if r.Buffer.Len() != 0 || r.Buffer.Cap() != bufferCap {
		t.Fatal("bytes.Buffer.Reset was not called")
	}
	if len(r.ArraySlices[0]) != 0 || &r.ArraySlices[0][:cap(r.ArraySlices[0])][0] != backing ||
		len(r.ArrayMaps[0]) != 0 || len(mapAlias) != 0 || r.ArrayMaps[0] == nil ||
		r.ArrayChildren[0].Name != "" || len(r.ArrayChildren[0].Values) != 0 ||
		r.ArrayChildren[0].Values == nil || r.ArrayPointers[0] != &number || number != 0 || r.Matrix != ([2][2]int{}) {
		t.Fatal("array elements were not reset according to their types")
	}
	if r.ArraySlices[1] != nil || r.ArrayMaps[1] != nil || r.ArrayPointers[1] != nil ||
		r.Any != nil || r.Handler != nil || r.Channel != nil || r.Interface != nil || r.Unsafe != nil {
		t.Fatal("nil or nullable fields were not reset correctly")
	}
}

func TestGenericReset(t *testing.T) {
	number := 7
	box := Box[int]{Value: 9, Pointer: &number, Items: []int{1, 2}, Child: &Box[int]{Value: 8}}
	backing, leaf := &box.Items[0], box.Child
	box.Reset()
	if box.Value != 0 || box.Pointer != &number || number != 0 || len(box.Items) != 0 ||
		&box.Items[:cap(box.Items)][0] != backing || box.Child != leaf || leaf.Value != 0 {
		t.Fatal("generic Reset did not preserve pointer or slice identity")
	}
	pair := Pair[string, int]{First: "first", Second: 8}
	pair.Reset()
	if pair.First != "" || pair.Second != 0 {
		t.Fatal("generic fields were not zeroed")
	}
	slice := Box[[]int]{Value: []int{1, 2, 3}}
	storage := &slice.Value[0]
	slice.Reset()
	if len(slice.Value) != 0 || &slice.Value[:cap(slice.Value)][0] != storage {
		t.Fatal("slice stored in a type parameter lost its storage")
	}
	entries := map[string]int{"key": 1}
	mapping := Box[map[string]int]{Value: entries}
	mapping.Reset()
	mapping.Value["new"] = 2
	if len(entries) != 1 || entries["new"] != 2 {
		t.Fatal("map stored in a type parameter was not cleared in place")
	}
	nested := &child.Child{Name: "generic child", Values: []int{1}}
	pointer := Box[*child.Child]{Value: nested}
	pointer.Reset()
	if pointer.Value != nested || nested.Name != "" || len(nested.Values) != 0 || nested.Values == nil {
		t.Fatal("pointer to a resettable structure in a type parameter was not reset")
	}
	array := Box[[2]int]{Value: [2]int{1, 2}}
	array.Reset()
	if array.Value != ([2]int{}) {
		t.Fatal("array stored in a type parameter was not reset")
	}
	plain := Box[Bare]{Value: Bare{Count: 1, Names: []string{"clear"}}}
	plain.Reset()
	if plain.Value.Count != 0 || plain.Value.Names != nil {
		t.Fatal("plain structure stored in a type parameter was not zeroed")
	}
}
