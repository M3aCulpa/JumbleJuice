package encoder

import (
	"fmt"
	"jumblejuice/internal/common"
	"sync"
)

// encoded data with metadata
type Encoded struct {
	Name     string // encoder name
	Chunks   []string // encoded payload chunks
	Size     int64  // original data size
	Checksum string // sha256 hex checksum
}

// interface for all encoders
type Encoder interface {
	Name() string
	Encode(data []byte) (Encoded, error)
	Decode(encoded Encoded) ([]byte, error)
}

// registry for encoders
var (
	encoders = make(map[string]Encoder)
	mu       sync.RWMutex
)

// adds an encoder to the registry
func Register(encoder Encoder) {
	mu.Lock()
	defer mu.Unlock()
	encoders[encoder.Name()] = encoder
}

// retrieves an encoder by name
func GetEncoder(name string) (Encoder, error) {
	mu.RLock()
	defer mu.RUnlock()

	encoder, ok := encoders[name]
	if !ok {
		return nil, fmt.Errorf("encoder not found: %s", name)
	}

	return encoder, nil
}

// returns all registered encoder names
func ListEncoders() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(encoders))
	for name := range encoders {
		names = append(names, name)
	}
	return names
}

// computes sha256 hex for use in encoders.
func Checksum(data []byte) string {
	return common.SHA256Hex(data)
}

// validates decoded data against expected checksum.
func verifyChecksum(data []byte, expected string) error {
	if expected == "" {
		return nil
	}
	actual := Checksum(data)
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}
