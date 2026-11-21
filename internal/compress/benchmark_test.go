package compress

import (
	"bytes"
	"fmt"
	"testing"
)

var benchmarkBytes []byte

func BenchmarkCompress(b *testing.B) {
	for _, size := range []int{1024, 4096, 65536} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			metric := []byte(`{"id":"Alloc","type":"gauge","value":123.5},`)
			data := bytes.Repeat(metric, (size+len(metric)-1)/len(metric))[:size]
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for b.Loop() {
				var err error
				benchmarkBytes, err = Compress(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecompress(b *testing.B) {
	data := bytes.Repeat([]byte(`{"id":"Alloc","type":"gauge","value":123.5},`), 100)
	compressed, err := Compress(data)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for b.Loop() {
		benchmarkBytes, err = Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}
