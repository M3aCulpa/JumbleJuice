package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadInput(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("Hello, JumbleJuice!")

	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	data, err := ReadInput(testFile)
	if err != nil {
		t.Fatalf("ReadInput() error = %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("ReadInput() = %v, want %v", data, testData)
	}

	_, err = ReadInput("nonexistent.txt")
	if err == nil {
		t.Error("ReadInput() should error on non-existent file")
	}
}

func TestSHA256Hex(t *testing.T) {
	hash := SHA256Hex([]byte("hello"))
	if len(hash) != 64 {
		t.Errorf("SHA256Hex() length = %d, want 64", len(hash))
	}

	// same input should produce same hash
	hash2 := SHA256Hex([]byte("hello"))
	if hash != hash2 {
		t.Error("SHA256Hex() not deterministic")
	}
}
