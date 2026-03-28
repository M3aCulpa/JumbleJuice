package encoder

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// encodes data as base64.
type Base64Encoder struct{}

func (b *Base64Encoder) Name() string { return "b64" }

func (b *Base64Encoder) Encode(data []byte) (Encoded, error) {
	if len(data) == 0 {
		return Encoded{Name: b.Name()}, nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	// split into 76-char lines (standard mime)
	var chunks []string
	const lineWidth = 76
	for i := 0; i < len(encoded); i += lineWidth {
		end := i + lineWidth
		if end > len(encoded) {
			end = len(encoded)
		}
		chunks = append(chunks, encoded[i:end])
	}

	return Encoded{
		Name:     b.Name(),
		Chunks:   chunks,
		Size:     int64(len(data)),
		Checksum: Checksum(data),
	}, nil
}

func (b *Base64Encoder) Decode(encoded Encoded) ([]byte, error) {
	if encoded.Name != b.Name() {
		return nil, fmt.Errorf("invalid encoder: expected %s, got %s", b.Name(), encoded.Name)
	}
	if len(encoded.Chunks) == 0 || encoded.Size == 0 {
		return []byte{}, nil
	}

	result, err := base64.StdEncoding.DecodeString(strings.Join(encoded.Chunks, ""))
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %w", err)
	}

	if err := verifyChecksum(result, encoded.Checksum); err != nil {
		return nil, err
	}
	return result, nil
}

func init() { Register(&Base64Encoder{}) }
