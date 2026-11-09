package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var writerPool = sync.Pool{
	New: func() any { return gzip.NewWriter(io.Discard) },
}

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := writerPool.Get().(*gzip.Writer)
	w.Reset(&b)
	defer func() {
		// Release the output buffer while retaining the compressor's internal working memory.
		w.Reset(io.Discard)
		writerPool.Put(w)
	}()
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

func NewWriter(w io.Writer) *gzip.Writer {
	return gzip.NewWriter(w)
}

func NewReader(r io.Reader) (*gzip.Reader, error) {
	return gzip.NewReader(r)
}
