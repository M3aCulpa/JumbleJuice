package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempInput(t *testing.T, content []byte) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(f, content, 0600); err != nil {
		t.Fatalf("failed to write test input: %v", err)
	}
	return f
}

func TestRun(t *testing.T) {
	code, err := run(writeTempInput(t, []byte("Hello, World!")), "ipv4", "c", false)
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
	code, err := run(writeTempInput(t, []byte("Hello, World!")), "ipv4", "c", true)
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	if strings.Contains(code, "=== decoder ===") {
		t.Error("raw should not contain decoder")
	}
}

func TestRunEmptyInput(t *testing.T) {
	code, err := run(writeTempInput(t, []byte{}), "ipv4", "c", false)
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	if !strings.Contains(code, "=== data ===") {
		t.Error("empty input should still produce data section")
	}
}

func TestRunInvalidEncoder(t *testing.T) {
	_, err := run(writeTempInput(t, []byte("test")), "invalid", "c", false)
	if err == nil {
		t.Error("expected error for invalid encoder")
	}
}

func TestRunInvalidLanguage(t *testing.T) {
	_, err := run(writeTempInput(t, []byte("test")), "ipv4", "invalid", false)
	if err == nil {
		t.Error("expected error for invalid language")
	}
}
