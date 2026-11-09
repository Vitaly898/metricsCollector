// Package compress предоставляет gzip-сжатие и распаковку тел HTTP-сообщений.
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

// Compress возвращает gzip-представление data с переиспользованием компрессоров.
// Результат принадлежит вызывающему коду и не меняется последующими вызовами.
// Функция безопасна для параллельных вызовов.
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

// Decompress распаковывает gzip-данные и возвращает ошибку повреждённого потока.
func Decompress(data []byte) ([]byte, error) {
	r, err := NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

// NewWriter создаёт gzip-писатель поверх w.
// Вызывающий код должен вызвать Close для завершения потока; сам w не закрывается.
func NewWriter(w io.Writer) *gzip.Writer {
	return gzip.NewWriter(w)
}

// NewReader создаёт gzip-читатель или возвращает ошибку заголовка потока.
// Вызывающий код закрывает читатель и отдельно управляет закрытием r.
func NewReader(r io.Reader) (*gzip.Reader, error) {
	return gzip.NewReader(r)
}
