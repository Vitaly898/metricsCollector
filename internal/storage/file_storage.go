package storage

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type FileStorage struct {
	*MemStorage
	filePath      string
	storeInterval int
	stop          chan struct{}
	done          chan struct{}
}

func NewFileStorage(mem *MemStorage, filePath string, storeInterval int) *FileStorage {
	return &FileStorage{
		MemStorage:    mem,
		filePath:      filePath,
		storeInterval: storeInterval,
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}
}

func (fs *FileStorage) UpdateGauge(name string, val float64) {
	fs.MemStorage.UpdateGauge(name, val)
	if fs.storeInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("Failed to save metrics: %v", err)
		}
	}
}

func (fs *FileStorage) UpdateCounter(name string, val int64) {
	fs.MemStorage.UpdateCounter(name, val)
	if fs.storeInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("Failed to save metrics: %v", err)
		}
	}
}

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

func (fs *FileStorage) Load() error {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metrics []struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}

	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				fs.gauge[m.ID] = *m.Value
			}
		case "counter":
			if m.Delta != nil {
				fs.counter[m.ID] = *m.Delta
			}
		}
	}

	return nil
}

func (fs *FileStorage) Start() {
	if fs.storeInterval <= 0 {
		return
	}
	go fs.saver()
}

func (fs *FileStorage) saver() {
	defer close(fs.done)
	ticker := time.NewTicker(time.Duration(fs.storeInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-fs.stop:
			if err := fs.Save(); err != nil {
				log.Printf("Failed to save metrics: %v", err)
			}
			return
		case <-ticker.C:
			if err := fs.Save(); err != nil {
				log.Printf("Failed to save metrics: %v", err)
			}
		}
	}
}

func (fs *FileStorage) Stop() {
	if fs.storeInterval <= 0 {
		return
	}
	close(fs.stop)
	<-fs.done
}
