package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileObserver записывает события аудита в файл по одному JSON-объекту на строку.
// Запись и закрытие защищены от одновременного доступа.
type FileObserver struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileObserver открывает path для дозаписи или создаёт файл с правами 0600.
// Вызывающий код отвечает за закрытие приёмника через Close.
func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &FileObserver{file: file}, nil
}

// Notify добавляет событие в конец файла, завершая запись переводом строки.
// Контекст не используется; ошибка сериализации или записи возвращается вызывающему коду.
func (o *FileObserver) Notify(_ context.Context, event Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := json.NewEncoder(o.file).Encode(event); err != nil {
		return fmt.Errorf("write audit event to file: %w", err)
	}
	return nil
}

// Close закрывает файл; последующие вызовы Notify возвращают ошибку записи.
func (o *FileObserver) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.file.Close()
}
