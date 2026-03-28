package encoder

import (
	"fmt"
	"strconv"
	"strings"
)

// encodes data as ipv4 addresses (4 bytes each).
type IPv4Encoder struct{}

func (e *IPv4Encoder) Name() string { return "ipv4" }

func (e *IPv4Encoder) Encode(data []byte) (Encoded, error) {
	if len(data) == 0 {
		return Encoded{Name: e.Name()}, nil
	}

	var chunks []string
	const chunkSize = 4

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := make([]byte, chunkSize)
		copy(chunk, data[i:end])

		chunks = append(chunks, fmt.Sprintf("%d.%d.%d.%d",
			chunk[0], chunk[1], chunk[2], chunk[3]))
	}

	return Encoded{
		Name:     e.Name(),
		Chunks:   chunks,
		Size:     int64(len(data)),
		Checksum: Checksum(data),
	}, nil
}

func (e *IPv4Encoder) Decode(encoded Encoded) ([]byte, error) {
	if encoded.Name != e.Name() {
		return nil, fmt.Errorf("invalid encoder: expected %s, got %s", e.Name(), encoded.Name)
	}

	var result []byte

	for _, addr := range encoded.Chunks {
		if err := ValidateIPv4(addr); err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		for _, part := range strings.Split(addr, ".") {
			octet, _ := strconv.Atoi(part)
			result = append(result, byte(octet))
		}
	}

	// trim padding
	if encoded.Size > 0 && encoded.Size < int64(len(result)) {
		result = result[:encoded.Size]
	}

	if err := verifyChecksum(result, encoded.Checksum); err != nil {
		return nil, err
	}
	return result, nil
}

func init() { Register(&IPv4Encoder{}) }
