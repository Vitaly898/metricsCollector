package hash

import (
	"bytes"
	"testing"
)

var benchmarkSignature string

func BenchmarkCompute(b *testing.B) {
	data := bytes.Repeat([]byte("metrics"), 600)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for b.Loop() {
		benchmarkSignature = Compute(data, "benchmark-key")
	}
}
