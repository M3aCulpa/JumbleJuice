package encoder

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// encodes data as hexadecimal byte values.
type HexEncoder struct{}

func (h *HexEncoder) Name() string { return "hex" }

func (h *HexEncoder) Encode(data []byte) (Encoded, error) {
	if len(data) == 0 {
		return Encoded{Name: h.Name()}, nil
	}

	var chunks []string
	const bytesPerLine = 16

	for i := 0; i < len(data); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		var hexBytes []string
		for _, b := range data[i:end] {
			hexBytes = append(hexBytes, fmt.Sprintf("0x%02X", b))
		}
		chunks = append(chunks, strings.Join(hexBytes, ", "))
	}

	return Encoded{
		Name:     h.Name(),
		Chunks:   chunks,
		Size:     int64(len(data)),
		Checksum: Checksum(data),
	}, nil
}

func (h *HexEncoder) Decode(encoded Encoded) ([]byte, error) {
	if encoded.Name != h.Name() {
		return nil, fmt.Errorf("invalid encoder: expected %s, got %s", h.Name(), encoded.Name)
	}
	if len(encoded.Chunks) == 0 || encoded.Size == 0 {
		return []byte{}, nil
	}

	result := make([]byte, 0, encoded.Size)

	for _, chunk := range encoded.Chunks {
		if chunk == "" {
			continue
		}
		for _, hexStr := range strings.Split(chunk, ",") {
			hexStr = strings.TrimSpace(hexStr)
			hexStr = strings.TrimPrefix(hexStr, "0x")
			hexStr = strings.TrimPrefix(hexStr, "0X")
			if hexStr == "" {
				continue
			}
			b, err := hex.DecodeString(hexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid hex value %q: %w", hexStr, err)
			}
			result = append(result, b...)
		}
	}

	if err := verifyChecksum(result, encoded.Checksum); err != nil {
		return nil, err
	}
	return result, nil
}

func init() { Register(&HexEncoder{}) }
