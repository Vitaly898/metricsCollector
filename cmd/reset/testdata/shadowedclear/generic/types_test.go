package generic

import "testing"

func TestResetWithTypeParameterNamedClear(t *testing.T) {
	value := Container[int]{Value: 8, Entries: map[string]int{"key": 1}}
	alias := value.Entries
	value.Reset()
	if value.Value != 0 || len(value.Entries) != 0 || len(alias) != 0 {
		t.Fatalf("generic reset did not clear the state: %+v", value)
	}
	value.Entries["new"] = 2
	if alias["new"] != 2 {
		t.Fatal("reset replaced the original map")
	}
}
