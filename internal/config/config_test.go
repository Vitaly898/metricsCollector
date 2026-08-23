package config

import (
	"os"
	"testing"
)

func TestParseDefaultValues(t *testing.T) {
	os.Args = []string{"server"}
	for _, key := range []string{"ADDRESS", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE"} {
		os.Unsetenv(key)
	}

	cfg := Parse()

	if cfg.Addr != "localhost:8080" {
		t.Errorf("Addr = %q, want localhost:8080", cfg.Addr)
	}
	if cfg.StoreInterval != 300 {
		t.Errorf("StoreInterval = %d, want 300", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "/tmp/metrics-db.json" {
		t.Errorf("FileStoragePath = %q, want /tmp/metrics-db.json", cfg.FileStoragePath)
	}
	if cfg.Restore != true {
		t.Errorf("Restore = %v, want true", cfg.Restore)
	}
}

func TestParseFlags(t *testing.T) {
	os.Args = []string{
		"server",
		"-a", "127.0.0.1:9090",
		"-i", "60",
		"-f", "/data/metrics.json",
		"-r=false",
	}
	for _, key := range []string{"ADDRESS", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE"} {
		os.Unsetenv(key)
	}

	cfg := Parse()

	if cfg.Addr != "127.0.0.1:9090" {
		t.Errorf("Addr = %q, want 127.0.0.1:9090", cfg.Addr)
	}
	if cfg.StoreInterval != 60 {
		t.Errorf("StoreInterval = %d, want 60", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "/data/metrics.json" {
		t.Errorf("FileStoragePath = %q, want /data/metrics.json", cfg.FileStoragePath)
	}
	if cfg.Restore != false {
		t.Errorf("Restore = %v, want false", cfg.Restore)
	}
}

func TestParseEnvOverridesFlags(t *testing.T) {
	os.Args = []string{
		"server",
		"-a", "127.0.0.1:9090",
		"-i", "60",
		"-f", "/data/metrics.json",
		"-r", "false",
	}
	t.Setenv("ADDRESS", "0.0.0.0:7777")
	t.Setenv("STORE_INTERVAL", "10")
	t.Setenv("FILE_STORAGE_PATH", "/env/metrics.json")
	t.Setenv("RESTORE", "true")

	cfg := Parse()

	if cfg.Addr != "0.0.0.0:7777" {
		t.Errorf("Addr = %q, want 0.0.0.0:7777", cfg.Addr)
	}
	if cfg.StoreInterval != 10 {
		t.Errorf("StoreInterval = %d, want 10", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "/env/metrics.json" {
		t.Errorf("FileStoragePath = %q, want /env/metrics.json", cfg.FileStoragePath)
	}
	if cfg.Restore != true {
		t.Errorf("Restore = %v, want true", cfg.Restore)
	}
}
