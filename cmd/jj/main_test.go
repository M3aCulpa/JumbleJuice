package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.bin")
	os.WriteFile(inputFile, []byte("Hello, World!"), 0644)

	code, err := run(inputFile, "ipv4", "c", false)
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	if !strings.Contains(code, "=== data ===") {
		t.Error("missing data marker")
	}
	if !strings.Contains(code, "=== decoder ===") {
		t.Error("missing decoder marker")
	}
}

func TestRunRaw(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.bin")
	os.WriteFile(inputFile, []byte("Hello, World!"), 0644)

	code, err := run(inputFile, "ipv4", "c", true)
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	if strings.Contains(code, "=== decoder ===") {
		t.Error("raw should not contain decoder")
	}
}

func TestRunEmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "empty.bin")
	os.WriteFile(inputFile, []byte{}, 0644)

	code, err := run(inputFile, "ipv4", "c", false)
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	if !strings.Contains(code, "=== data ===") {
		t.Error("empty input should still produce data section")
	}
}

func TestRunInvalidEncoder(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.bin")
	os.WriteFile(inputFile, []byte("test"), 0644)

	_, err := run(inputFile, "invalid", "c", false)
	if err == nil {
		t.Error("expected error for invalid encoder")
	}
}

func TestRunInvalidLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.bin")
	os.WriteFile(inputFile, []byte("test"), 0644)

	_, err := run(inputFile, "ipv4", "invalid", false)
	if err == nil {
		t.Error("expected error for invalid language")
	}
}
