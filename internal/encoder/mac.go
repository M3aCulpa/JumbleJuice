package encoder

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// encodes data as mac addresses (6 bytes each).
type MACEncoder struct{}

func (m *MACEncoder) Name() string { return "mac" }

func (m *MACEncoder) Encode(data []byte) (Encoded, error) {
	if len(data) == 0 {
		return Encoded{Name: m.Name()}, nil
	}

	var chunks []string
	const chunkSize = 6

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := make([]byte, chunkSize)
		copy(chunk, data[i:end])

		var parts []string
		for _, b := range chunk {
			parts = append(parts, fmt.Sprintf("%02X", b))
		}
		chunks = append(chunks, strings.Join(parts, ":"))
	}

	return Encoded{
		Name:     m.Name(),
		Chunks:   chunks,
		Size:     int64(len(data)),
		Checksum: Checksum(data),
	}, nil
}

func (m *MACEncoder) Decode(encoded Encoded) ([]byte, error) {
	if encoded.Name != m.Name() {
		return nil, fmt.Errorf("invalid encoder: expected %s, got %s", m.Name(), encoded.Name)
	}

	var result []byte

	for _, addr := range encoded.Chunks {
		if err := ValidateMAC(addr); err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		cleaned := strings.ReplaceAll(addr, ":", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		b, err := hex.DecodeString(cleaned)
		if err != nil {
			return nil, fmt.Errorf("invalid hex in MAC address %s: %w", addr, err)
		}
		result = append(result, b...)
	}

	if encoded.Size > 0 && encoded.Size < int64(len(result)) {
		result = result[:encoded.Size]
	}

	if err := verifyChecksum(result, encoded.Checksum); err != nil {
		return nil, err
	}
	return result, nil
}

func init() { Register(&MACEncoder{}) }
