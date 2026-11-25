package shadowedclear

import "testing"

func TestResetClearsMapsDespiteShadowing(t *testing.T) {
	pointed := map[string]int{"pointed": 3}
	value := Container{
		Values: map[string]int{"first": 1}, Labels: Labels{"second": 2},
		Pointer: &pointed, Array: [1]map[string]int{{"fourth": 4}},
	}
	values, labels, arrayMap := value.Values, value.Labels, value.Array[0]
	value.Reset()
	if len(value.Values) != 0 || len(value.Labels) != 0 || len(pointed) != 0 || len(value.Array[0]) != 0 {
		t.Fatalf("maps were not cleared: %+v; pointed: %v", value, pointed)
	}
	if customClearCalls != 0 {
		t.Fatalf("generated Reset called the user clear function %d times", customClearCalls)
	}
	value.Values["new"] = 5
	value.Labels["new"] = 6
	pointed["new"] = 7
	value.Array[0]["new"] = 8
	if values["new"] != 5 || labels["new"] != 6 || (*value.Pointer)["new"] != 7 || arrayMap["new"] != 8 || value.Pointer != &pointed {
		t.Fatal("reset replaced a map or a pointer")
	}
	if value.NilMap != nil {
		t.Fatal("reset allocated a nil map")
	}
}
