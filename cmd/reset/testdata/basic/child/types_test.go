package child

import "testing"

func TestReset(t *testing.T) {
	child := Child{Name: "child", Values: []int{1, 2}}
	child.Reset()
	if child.Name != "" || len(child.Values) != 0 || child.Values == nil {
		t.Fatalf("unexpected child after reset: %+v", child)
	}
}
