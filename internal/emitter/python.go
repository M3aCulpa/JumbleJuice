package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"strings"
)

type PythonEmitter struct{}

func (p *PythonEmitter) Language() string { return "python" }

func (p *PythonEmitter) Emit(encoded encoder.Encoded, raw bool) (string, error) {
	if raw {
		return p.data(encoded), nil
	}
	var sb strings.Builder
	sb.WriteString("# === imports ===\n")
	sb.WriteString(p.imports(encoded))
	sb.WriteString("\n# === data ===\n")
	sb.WriteString(p.data(encoded))
	sb.WriteString("\n# === decoder ===\n")
	sb.WriteString(p.decoder(encoded))
	return sb.String(), nil
}

func (p *PythonEmitter) imports(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "b64":
		return "import base64\n"
	default:
		return "# (no imports required)\n"
	}
}

func (p *PythonEmitter) data(encoded encoder.Encoded) string {
	var sb strings.Builder

	switch encoded.Name {
	case "hex", "dec":
		sb.WriteString("data = bytes([\n")
		for i, chunk := range encoded.Chunks {
			if i > 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString("    ")
			sb.WriteString(chunk)
		}
		sb.WriteString("\n])\n")

	case "b64":
		sb.WriteString("data = (\n")
		for _, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"\n", chunk))
		}
		sb.WriteString(")\n")

	default:
		// ipv4, ipv6, mac
		sb.WriteString("data = [\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("]\n")
	}

	sb.WriteString(fmt.Sprintf("data_original_size = %d\n", encoded.Size))
	return sb.String()
}

func (p *PythonEmitter) decoder(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "hex", "dec":
		return `def decode():
    return data[:data_original_size]
`
	case "b64":
		return `def decode():
    return base64.b64decode(data)[:data_original_size]
`
	case "ipv4":
		return `def decode():
    result = bytearray()
    for addr in data:
        for octet in addr.split("."):
            result.append(int(octet))
    return bytes(result[:data_original_size])
`
	case "ipv6", "mac":
		// both strip colons and parse hex
		return `def decode():
    result = bytearray()
    for addr in data:
        cleaned = addr.replace(":", "")
        result.extend(bytes.fromhex(cleaned))
    return bytes(result[:data_original_size])
`
	default:
		return fmt.Sprintf("# decoder for %s not implemented\n", encoded.Name)
	}
}

func init() { Register(&PythonEmitter{}) }
