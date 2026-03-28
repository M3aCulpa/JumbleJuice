package encoder

import (
	"fmt"
	"strconv"
	"strings"
)

// encodes data as decimal byte values.
type DecimalEncoder struct{}

func (d *DecimalEncoder) Name() string { return "dec" }

func (d *DecimalEncoder) Encode(data []byte) (Encoded, error) {
	if len(data) == 0 {
		return Encoded{Name: d.Name()}, nil
	}

	var chunks []string
	const bytesPerLine = 16

	for i := 0; i < len(data); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		var values []string
		for _, b := range data[i:end] {
			values = append(values, strconv.Itoa(int(b)))
		}
		chunks = append(chunks, strings.Join(values, ", "))
	}

	return Encoded{
		Name:     d.Name(),
		Chunks:   chunks,
		Size:     int64(len(data)),
		Checksum: Checksum(data),
	}, nil
}

func (d *DecimalEncoder) Decode(encoded Encoded) ([]byte, error) {
	if encoded.Name != d.Name() {
		return nil, fmt.Errorf("invalid encoder: expected %s, got %s", d.Name(), encoded.Name)
	}
	if len(encoded.Chunks) == 0 || encoded.Size == 0 {
		return []byte{}, nil
	}

	result := make([]byte, 0, encoded.Size)

	for _, chunk := range encoded.Chunks {
		if chunk == "" {
			continue
		}
		for _, val := range strings.Split(chunk, ",") {
			val = strings.TrimSpace(val)
			if val == "" {
				continue
			}
			num, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("invalid decimal value %q: %w", val, err)
			}
			if num < 0 || num > 255 {
				return nil, fmt.Errorf("decimal value %d out of byte range [0-255]", num)
			}
			result = append(result, byte(num))
		}
	}

	if err := verifyChecksum(result, encoded.Checksum); err != nil {
		return nil, err
	}
	return result, nil
}

func init() { Register(&DecimalEncoder{}) }
