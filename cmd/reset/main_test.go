package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExitCodes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), []byte("module example.com/cli\n\ngo 1.26.0\n"))
	writeFile(t, filepath.Join(dir, "types.go"), []byte("package cli\ntype Item struct { Value int }\n"))
	for _, tt := range []struct {
		name    string
		args    []string
		code    int
		message string
	}{
		{"success", []string{"-dir", dir}, 0, ""},
		{"help", []string{"-h"}, 0, "Usage of reset"},
		{"unknown flag", []string{"-unknown"}, 2, "flag provided but not defined"},
		{"extra argument", []string{"extra"}, 2, "usage: reset"},
		{"generation error", []string{"-dir", filepath.Join(dir, "missing")}, 1, "missing"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if got := run(tt.args, &stderr); got != tt.code {
				t.Fatalf("exit code = %d, want %d; stderr: %s", got, tt.code, stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.message) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.message)
			}
		})
	}
}
