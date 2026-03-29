package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"strings"
)

type GoEmitter struct{}

func (g *GoEmitter) Language() string { return "go" }

func (g *GoEmitter) Emit(encoded encoder.Encoded, raw bool) (string, error) {
	if raw {
		return g.data(encoded), nil
	}
	var sb strings.Builder
	sb.WriteString("// === imports ===\n")
	sb.WriteString(g.imports(encoded))
	sb.WriteString("\n// === data ===\n")
	sb.WriteString(g.data(encoded))
	sb.WriteString("\n// === decoder ===\n")
	sb.WriteString(g.decoder(encoded))
	return sb.String(), nil
}

func (g *GoEmitter) imports(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "b64":
		return "import \"encoding/base64\"\n"
	case "ipv4":
		return "import (\n\t\"strconv\"\n\t\"strings\"\n)\n"
	case "ipv6", "mac":
		return "import (\n\t\"encoding/hex\"\n\t\"strings\"\n)\n"
	default:
		return "// (no imports required)\n"
	}
}

func (g *GoEmitter) data(encoded encoder.Encoded) string {
	var sb strings.Builder

	switch encoded.Name {
	case "hex", "dec":
		sb.WriteString("var data = []byte{\n")
		for i, chunk := range encoded.Chunks {
			if i > 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString("\t")
			sb.WriteString(chunk)
		}
		sb.WriteString(",\n}\n")

	case "b64":
		sb.WriteString("var data = `")
		for _, chunk := range encoded.Chunks {
			sb.WriteString(chunk)
		}
		sb.WriteString("`\n")

	default:
		// ipv4, ipv6, mac
		sb.WriteString("var data = []string{\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("\t\"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("}\n")
	}

	sb.WriteString(fmt.Sprintf("const dataOriginalSize = %d\n", encoded.Size))
	return sb.String()
}

func (g *GoEmitter) decoder(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "hex", "dec":
		return `func decode() []byte {
	if len(data) < dataOriginalSize {
		return data
	}
	return data[:dataOriginalSize]
}
`
	case "b64":
		return `func decode() []byte {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic("failed to decode base64: " + err.Error())
	}
	return decoded[:dataOriginalSize]
}
`
	case "ipv4":
		return `func decode() []byte {
	result := make([]byte, 0, len(data)*4)
	for _, addr := range data {
		for _, part := range strings.Split(addr, ".") {
			b, err := strconv.Atoi(part)
			if err != nil {
				panic("invalid IPv4 octet: " + part)
			}
			result = append(result, byte(b))
		}
	}
	return result[:dataOriginalSize]
}
`
	case "ipv6", "mac":
		return `func decode() []byte {
	result := make([]byte, 0, len(data)*16)
	for _, addr := range data {
		cleaned := strings.ReplaceAll(addr, ":", "")
		b, err := hex.DecodeString(cleaned)
		if err != nil {
			panic("invalid hex in address: " + addr)
		}
		result = append(result, b...)
	}
	return result[:dataOriginalSize]
}
`
	default:
		return fmt.Sprintf("// decoder for %s not implemented\n", encoded.Name)
	}
}

func init() { Register(&GoEmitter{}) }
