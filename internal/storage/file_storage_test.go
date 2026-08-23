package storage

import (
	"os"
	"testing"
	"time"
)

func TestFileStorageSaveAndLoad(t *testing.T) {
	path := "/tmp/test-metrics-save-load.json"
	defer os.Remove(path)

	mem := NewMemStorage()
	fs := NewFileStorage(mem, path, 60, false)

	mem.UpdateGauge("Alloc", 123.45)
	mem.UpdateCounter("PollCount", 5)

	if err := fs.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	newMem := NewMemStorage()
	newFs := NewFileStorage(newMem, path, 60, false)
	if err := newFs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	v, ok := newMem.GetGauge("Alloc")
	if !ok || v != 123.45 {
		t.Errorf("Alloc = %v, want 123.45", v)
	}

	c, ok := newMem.GetCounter("PollCount")
	if !ok || c != 5 {
		t.Errorf("PollCount = %v, want 5", c)
	}
}

func TestFileStorageSyncSaveOnUpdate(t *testing.T) {
	path := "/tmp/test-metrics-sync.json"
	defer os.Remove(path)

	mem := NewMemStorage()
	fs := NewFileStorage(mem, path, 0, false)

	fs.UpdateGauge("Alloc", 99.9)

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file was not created synchronously: %v", err)
	}

	newMem := NewMemStorage()
	newFs := NewFileStorage(newMem, path, 0, false)
	if err := newFs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	v, ok := newMem.GetGauge("Alloc")
	if !ok || v != 99.9 {
		t.Errorf("Alloc = %v, want 99.9", v)
	}
}

func TestFileStoragePeriodicSave(t *testing.T) {
	path := "/tmp/test-metrics-periodic.json"
	defer os.Remove(path)

	mem := NewMemStorage()
	fs := NewFileStorage(mem, path, 1, false)
	fs.storeInterval = 10 * time.Millisecond
	fs.Start()
	defer fs.Stop()

	mem.UpdateGauge("Alloc", 77.7)

	deadline := time.Now().Add(2 * time.Second)
	for {
		newMem := NewMemStorage()
		newFs := NewFileStorage(newMem, path, 1, false)
		if err := newFs.Load(); err == nil {
			if v, ok := newMem.GetGauge("Alloc"); ok && v == 77.7 {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("metric was not saved periodically")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestFileStorageLoadMissingFile(t *testing.T) {
	path := "/tmp/test-metrics-missing.json"
	defer os.Remove(path)

	mem := NewMemStorage()
	fs := NewFileStorage(mem, path, 60, false)

	if err := fs.Load(); err != nil {
		t.Errorf("Load should not return error for missing file: %v", err)
	}
}
