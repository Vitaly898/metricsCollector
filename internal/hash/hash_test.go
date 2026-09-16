package hash

import (
	"testing"
)

func TestHash(t *testing.T) {
	sum1 := Compute([]byte("body"), "secret")
	sum2 := Compute([]byte("body"), "secret")
	if sum1 != sum2 {
		t.Error("hash is not deterministic")
	}
	if Compute([]byte("body"), "secret") == Compute([]byte("body"), "otherKey") {
		t.Error("hash doesn't depend on key")
	}
	if Compute([]byte("body"), "secret") == Compute([]byte("body2"), "secret") {
		t.Error("hash doesn't depend on data")
	}
	sum := Compute([]byte("The quick brown fox jumps over the lazy dog"), "key")
	expected := "97yD9DBThCSxMpjmqm+xQ+9NWaFJRhdZl0edvC0aPNg="
	if sum != expected {
		t.Errorf("got %q, want %q", sum, expected)
	}
}
