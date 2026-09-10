package hash

import (
	"testing"
)

func TestHash(t *testing.T) {
	if Compute([]byte("body"), "secret") != Compute([]byte("body"), "secret") {
		t.Error("hash is not deterministic")
	}
	if Compute([]byte("body"), "secret") == Compute([]byte("body"), "otherKey") {
		t.Error("hash doesn't depend on key")
	}
	if Compute([]byte("body"), "secret") == Compute([]byte("body2"), "secret") {
		t.Error("hash doesn't depend on data")
	}
	sum := Compute([]byte("The quick brown fox jumps over the lazy dog"), "key")
	expected := "51729876100348eb46ed8c4bf39efa4037a3a2c687f864348ed69292a67ffdbc"
	if sum != expected {
		t.Errorf("got %q, want %q", sum, expected)
	}
}
