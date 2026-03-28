package encoder

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// encodes data as ipv6 addresses (16 bytes each).
type IPv6Encoder struct{}

func (e *IPv6Encoder) Name() string { return "ipv6" }

func (e *IPv6Encoder) Encode(data []byte) (Encoded, error) {
	if len(data) == 0 {
		return Encoded{Name: e.Name()}, nil
	}

	var chunks []string
	const chunkSize = 16

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := make([]byte, chunkSize)
		copy(chunk, data[i:end])

		ip := net.IP(chunk)
		addr := ip.String()

		// net.IP may produce ipv4-mapped notation; fall back to manual formatting
		if addr == "" || strings.Contains(addr, ".") {
			var parts []string
			for j := 0; j < 16; j += 2 {
				part := fmt.Sprintf("%02x%02x", chunk[j], chunk[j+1])
				part = strings.TrimLeft(part, "0")
				if part == "" {
					part = "0"
				}
				parts = append(parts, part)
			}
			addr = compressIPv6(strings.Join(parts, ":"))
		}

		chunks = append(chunks, addr)
	}

	return Encoded{
		Name:     e.Name(),
		Chunks:   chunks,
		Size:     int64(len(data)),
		Checksum: Checksum(data),
	}, nil
}

func (e *IPv6Encoder) Decode(encoded Encoded) ([]byte, error) {
	if encoded.Name != e.Name() {
		return nil, fmt.Errorf("invalid encoder: expected %s, got %s", e.Name(), encoded.Name)
	}

	var result []byte

	for _, addr := range encoded.Chunks {
		expanded := expandIPv6(addr)
		parts := strings.Split(expanded, ":")
		if len(parts) != 8 {
			return nil, fmt.Errorf("invalid IPv6 address after expansion: %s -> %s", addr, expanded)
		}
		for _, part := range parts {
			for len(part) < 4 {
				part = "0" + part
			}
			b, err := hex.DecodeString(part)
			if err != nil {
				return nil, fmt.Errorf("invalid hex group %q in address %s: %w", part, addr, err)
			}
			result = append(result, b...)
		}
	}

	if encoded.Size > 0 && encoded.Size < int64(len(result)) {
		result = result[:encoded.Size]
	}

	if err := verifyChecksum(result, encoded.Checksum); err != nil {
		return nil, err
	}
	return result, nil
}

// replaces the longest run of consecutive "0" groups with "::".
func compressIPv6(addr string) string {
	parts := strings.Split(addr, ":")

	maxStart, maxLen := -1, 0
	curStart, curLen := -1, 0

	for i, part := range parts {
		if part == "0" {
			if curStart == -1 {
				curStart = i
				curLen = 1
			} else {
				curLen++
			}
			if curLen > maxLen {
				maxStart = curStart
				maxLen = curLen
			}
		} else {
			curStart = -1
			curLen = 0
		}
	}

	if maxLen < 2 {
		return addr
	}
	if maxLen == len(parts) {
		return "::"
	}

	left := parts[:maxStart]
	right := parts[maxStart+maxLen:]

	var sb strings.Builder
	sb.WriteString(strings.Join(left, ":"))
	sb.WriteString("::")
	sb.WriteString(strings.Join(right, ":"))
	return sb.String()
}

// expands compressed ipv6 notation to 8 colon-separated groups.
func expandIPv6(addr string) string {
	if !strings.Contains(addr, "::") {
		return addr
	}

	halves := strings.Split(addr, "::")
	left := splitNonEmpty(halves[0], ":")
	right := []string{}
	if len(halves) > 1 && halves[1] != "" {
		right = splitNonEmpty(halves[1], ":")
	}

	zerosNeeded := 8 - len(left) - len(right)

	var expanded []string
	expanded = append(expanded, left...)
	for i := 0; i < zerosNeeded; i++ {
		expanded = append(expanded, "0")
	}
	expanded = append(expanded, right...)

	return strings.Join(expanded, ":")
}

func splitNonEmpty(s, sep string) []string {
	var result []string
	for _, p := range strings.Split(s, sep) {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() { Register(&IPv6Encoder{}) }
