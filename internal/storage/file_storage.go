package storage

import (
	"encoding/json"
	"os"
	"time"

	"go.uber.org/zap"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

var logger = zap.Must(zap.NewProduction()).Sugar()

// FileStorage дополняет MemStorage сохранением метрик в JSON-файл.
// Изменения сначала применяются в памяти; ошибка записи файла не откатывает их.
type FileStorage struct {
	// MemStorage предоставляет операции чтения и хранение текущих значений.
	*MemStorage
	filePath      string
	storeInterval time.Duration
	stop          chan struct{}
	done          chan struct{}
}

// NewFileStorage оборачивает mem файловым хранилищем с интервалом в секундах.
// При restore значения загружаются из filePath; ошибки загрузки записываются в лог.
// Нулевой интервал включает синхронную запись, положительный требует вызова Start.
func NewFileStorage(mem *MemStorage, filePath string, storeInterval int, restore bool) *FileStorage {
	fs := &FileStorage{
		MemStorage:    mem,
		filePath:      filePath,
		storeInterval: time.Duration(storeInterval) * time.Second,
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}
	if restore {
		if err := fs.Load(); err != nil {
			logger.Errorw("Failed to load metrics", "error", err)
		}
	}
	return fs
}

// UpdateGauge заменяет gauge в памяти и при нулевом интервале сохраняет файл.
// Ошибка синхронного сохранения возвращается вызывающему коду.
func (fs *FileStorage) UpdateGauge(name string, val float64) error {
	if err := fs.MemStorage.UpdateGauge(name, val); err != nil {
		return err
	}
	if fs.storeInterval == 0 {
		if err := fs.Save(); err != nil {
			return err
		}
	}
	return nil
}

// UpdateCounter увеличивает counter в памяти и при нулевом интервале сохраняет файл.
// Ошибка синхронного сохранения возвращается вызывающему коду.
func (fs *FileStorage) UpdateCounter(name string, val int64) error {
	if err := fs.MemStorage.UpdateCounter(name, val); err != nil {
		return err
	}
	if fs.storeInterval == 0 {
		if err := fs.Save(); err != nil {
			return err
		}
	}
	return nil
}

// UpdateMetrics обновляет пакет в памяти и при нулевом интервале сохраняет файл.
// Ошибки сохранения файла записываются в лог, но не возвращаются.
func (fs *FileStorage) UpdateMetrics(metrics []models.Metrics) error {
	if err := fs.MemStorage.UpdateMetrics(metrics); err != nil {
		return err
	}
	if fs.storeInterval == 0 {
		if err := fs.Save(); err != nil {
			logger.Errorw("Failed to save metrics", "error", err)
		}
	}
	return nil
}

// Save записывает снимок всех метрик во временный файл и переименовывает его в filePath.
func (fs *FileStorage) Save() error {
	fs.mu.Lock()
	metrics := fs.getAllMetricsLocked()
	fs.mu.Unlock()

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := fs.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fs.filePath)
}

// Load восстанавливает значения из JSON-файла, заменяя одноимённые метрики в памяти.
// Отсутствующий файл не считается ошибкой.
func (fs *FileStorage) Load() error {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metrics []models.Metrics

	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	for _, m := range metrics {
		switch m.MType {
		case models.Gauge:
			if m.Value != nil {
				fs.gauge[m.ID] = *m.Value
			}
		case models.Counter:
			if m.Delta != nil {
				fs.counter[m.ID] = *m.Delta
			}
		}
	}

	return nil
}

// Start запускает периодическое сохранение при положительном интервале.
// Для экземпляра с фоновым сохранением вызывается один раз перед Stop.
func (fs *FileStorage) Start() {
	if fs.storeInterval <= 0 {
		return
	}
	go fs.saver()
}

func (fs *FileStorage) saver() {
	defer close(fs.done)
	ticker := time.NewTicker(fs.storeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-fs.stop:
			if err := fs.Save(); err != nil {
				logger.Errorw("Failed to save metrics", "error", err)
			}
			return
		case <-ticker.C:
			if err := fs.Save(); err != nil {
				logger.Errorw("Failed to save metrics", "error", err)
			}
		}
	}
}

// Stop останавливает запущенное фоновое сохранение и ожидает финальной записи.
// При положительном интервале вызывается один раз после Start;
// при неположительном интервале ничего не делает.
func (fs *FileStorage) Stop() {
	if fs.storeInterval <= 0 {
		return
	}
	close(fs.stop)
	<-fs.done
}
