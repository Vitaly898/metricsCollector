package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	for _, fixture := range []string{"basic", "advanced"} {
		t.Run(fixture, func(t *testing.T) { testGenerate(t, fixture) })
	}
}

func testGenerate(t *testing.T, fixture string) {
	t.Helper()
	dir := fixtureModule(t, fixture)
	if err := generate(dir); err != nil {
		t.Fatal(err)
	}
	rootOutput := readFile(t, filepath.Join(dir, generatedFile))
	childOutput := readFile(t, filepath.Join(dir, "child", generatedFile))
	if bytes.Contains(rootOutput, []byte("Unmarked")) || bytes.Contains(rootOutput, []byte("TestOnly")) {
		t.Fatal("generated methods for an unmarked or test-only structure")
	}
	if _, err := os.Stat(filepath.Join(dir, "plain", generatedFile)); !os.IsNotExist(err) {
		t.Fatalf("unexpected generated file for an unmarked package: %v", err)
	}
	cmd := exec.Command("go", "test", "-race", "./...")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated code failed its behavior tests: %v\n%s", err, output)
	}
	if err := generate(dir); err != nil {
		t.Fatalf("second generation: %v", err)
	}
	if !bytes.Equal(rootOutput, readFile(t, filepath.Join(dir, generatedFile))) ||
		!bytes.Equal(childOutput, readFile(t, filepath.Join(dir, "child", generatedFile))) {
		t.Fatal("repeated generation changed the output")
	}
}

func TestGenerateRejectsInvalidInputWithoutWriting(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"not a struct", "// generate:reset\ntype Unsupported int", "generate:reset requires a struct type"},
		{"alias", "// generate:reset\ntype Unsupported = struct{}", "generate:reset requires a defined struct type"},
		{"field conflict", "// generate:reset\ntype Unsupported struct { Reset int }", "field conflicts with the generated method"},
		{"existing method", "// generate:reset\ntype Unsupported struct{}\nfunc (*Unsupported) Reset(int) {}", "existing Reset must have no parameters or results"},
		{"broken source", "type Unsupported struct {", "expected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := fixtureModule(t)
			writeFile(t, filepath.Join(dir, "invalid", "types.go"), []byte("package invalid\n\n"+tt.source+"\n"))
			err := generate(dir)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("generate error = %v, want %q", err, tt.want)
			}
			if _, err := os.Stat(filepath.Join(dir, generatedFile)); !os.IsNotExist(err) {
				t.Fatalf("generation modified files before validation finished: %v", err)
			}
		})
	}
}

func TestGeneratePreservesHandwrittenFile(t *testing.T) {
	dir := fixtureModule(t)
	path := filepath.Join(dir, generatedFile)
	original := []byte("package sample\n\n// This file belongs to the user.\n")
	writeFile(t, path, original)
	if err := generate(dir); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("generate error = %v, want refusal to overwrite", err)
	}
	if !bytes.Equal(original, readFile(t, path)) {
		t.Fatal("handwritten file was modified")
	}
}

func TestGenerateWithNoMarkers(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), []byte("module example.com/plain\n\ngo 1.26.0\n"))
	writeFile(t, filepath.Join(dir, "types.go"), []byte("package plain\n\ntype Plain struct { Value int }\n"))
	if err := generate(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, generatedFile)); !os.IsNotExist(err) {
		t.Fatalf("unexpected generated file: %v", err)
	}
}

func TestRegenerateAfterSourceChanges(t *testing.T) {
	t.Setenv("GOWORK", "off")
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), []byte("module example.com/changes\n\ngo 1.26.0\n"))
	path := filepath.Join(dir, "types.go")
	writeFile(t, path, []byte("package changes\n// generate:reset\ntype Item struct { Old string }\nfunc use(item *Item) { item.Reset() }\n"))
	if err := generate(dir); err != nil {
		t.Fatalf("initial generation with a call to a future method: %v", err)
	}
	writeFile(t, path, []byte("package changes\n// generate:reset\ntype Item struct { New int }\nfunc use(item *Item) { item.Reset() }\n"))
	if err := generate(dir); err != nil {
		t.Fatalf("regeneration after a field was removed: %v", err)
	}
	data := readFile(t, filepath.Join(dir, generatedFile))
	if bytes.Contains(data, []byte(".Old")) || !bytes.Contains(data, []byte(".New = 0")) {
		t.Fatalf("generated code was not updated:\n%s", data)
	}
	writeFile(t, path, []byte("package changes\ntype Replacement struct { Value int }\n"))
	if err := generate(dir); err != nil {
		t.Fatalf("regeneration after a marked type was removed: %v", err)
	}
	if bytes.Contains(readFile(t, filepath.Join(dir, generatedFile)), []byte("func ")) {
		t.Fatal("obsolete methods were not removed")
	}
}

func fixtureModule(t *testing.T, fixtures ...string) string {
	t.Helper()
	t.Setenv("GOWORK", "off")
	t.Setenv("GOPROXY", "off")
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), []byte("module example.com/sample\n\ngo 1.26.0\n"))
	fixture := "basic"
	if len(fixtures) > 0 {
		fixture = fixtures[0]
	}
	root := filepath.Join("testdata", fixture)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		writeFile(t, filepath.Join(dir, relative), readFile(t, path))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
