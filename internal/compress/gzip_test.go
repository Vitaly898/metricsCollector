package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
	"testing"
)

func decodeGzip(t *testing.T, compressed []byte) []byte {
	t.Helper()
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCompressRoundTrip(t *testing.T) {
	for _, data := range [][]byte{nil, {}, []byte("metrics"), bytes.Repeat([]byte("Alloc"), 20000)} {
		compressed, err := Compress(data)
		if err != nil {
			t.Fatal(err)
		}
		if got := decodeGzip(t, compressed); !bytes.Equal(got, data) {
			t.Errorf("decoded data differs for input of %d bytes", len(data))
		}
		got, err := Decompress(compressed)
		if err != nil || !bytes.Equal(got, data) {
			t.Errorf("Decompress: error %v, data matches: %v", err, bytes.Equal(got, data))
		}
	}
}

func TestCompressResultsRemainIndependent(t *testing.T) {
	first := bytes.Repeat([]byte("first metric "), 1000)
	compressed, err := Compress(first)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := bytes.Clone(compressed)
	for range 10 {
		if _, err := Compress(bytes.Repeat([]byte("other metric "), 2000)); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(compressed, snapshot) {
		t.Fatal("reusing the compressor changed a previous result")
	}
	if !bytes.Equal(decodeGzip(t, compressed), first) {
		t.Fatal("previous compressed result is corrupted")
	}
}

func TestCompressConcurrentCalls(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 32 {
		wg.Go(func() {
			data := bytes.Repeat([]byte{byte(i)}, 10000+i)
			for range 5 {
				compressed, err := Compress(data)
				if err != nil {
					t.Error(err)
					return
				}
				// Use the standard reader so this test does not share compressor state.
				reader, err := gzip.NewReader(bytes.NewReader(compressed))
				if err != nil {
					t.Error(err)
					return
				}
				got, readErr := io.ReadAll(reader)
				closeErr := reader.Close()
				if readErr != nil || closeErr != nil || !bytes.Equal(got, data) {
					t.Errorf("concurrent gzip result corrupted: read %v, close %v", readErr, closeErr)
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestDecompressRejectsInvalidData(t *testing.T) {
	compressed, err := Compress([]byte("metrics"))
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{[]byte("invalid gzip"), compressed[:len(compressed)-1]} {
		if _, err := Decompress(data); err == nil {
			t.Error("expected an error for invalid or truncated gzip")
		}
	}
}
