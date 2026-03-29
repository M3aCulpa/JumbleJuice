package common

import (
	"fmt"
	"io"
	"os"
)

// max payload size (100 MB) covers all realistic shellcode sizes.
const MaxPayloadSize = 100 * 1024 * 1024

// reads data from a file path, or stdin if path is "" or "-".
func ReadInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, MaxPayloadSize+1))
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
		if len(data) > MaxPayloadSize {
			return nil, fmt.Errorf("input exceeds maximum size (%d bytes)", MaxPayloadSize)
		}
		return data, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	if info.Size() > MaxPayloadSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), MaxPayloadSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}
