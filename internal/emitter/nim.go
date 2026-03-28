package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"strings"
)

type NimEmitter struct{}

func (n *NimEmitter) Language() string { return "nim" }

func (n *NimEmitter) Emit(encoded encoder.Encoded, raw bool) (string, error) {
	if raw {
		return n.data(encoded), nil
	}
	var sb strings.Builder
	sb.WriteString("# === imports ===\n")
	sb.WriteString(n.imports(encoded))
	sb.WriteString("\n# === data ===\n")
	sb.WriteString(n.data(encoded))
	sb.WriteString("\n# === decoder ===\n")
	sb.WriteString(n.decoder(encoded))
	return sb.String(), nil
}

func (n *NimEmitter) imports(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "b64":
		return "import base64\n"
	case "ipv4", "ipv6", "mac":
		return "import strutils\n"
	default:
		return "# (no imports required)\n"
	}
}

func (n *NimEmitter) data(encoded encoder.Encoded) string {
	var sb strings.Builder

	switch encoded.Name {
	case "hex", "dec":
		sb.WriteString("let data: seq[byte] = @[\n")
		for i, chunk := range encoded.Chunks {
			if i > 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString("    ")
			sb.WriteString(chunk)
		}
		sb.WriteString("\n]\n")

	case "b64":
		sb.WriteString("let data =\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(" &")
			}
			sb.WriteString("\n")
		}

	default:
		// ipv4, ipv6, mac
		sb.WriteString("let data = @[\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("]\n")
	}

	sb.WriteString(fmt.Sprintf("const dataOriginalSize = %d\n", encoded.Size))
	return sb.String()
}

func (n *NimEmitter) decoder(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "hex", "dec":
		return `proc decode(): seq[byte] =
    data[0 ..< dataOriginalSize]
`
	case "b64":
		return `proc decode(): seq[byte] =
    let decoded = base64.decode(data)
    cast[seq[byte]](decoded[0 ..< dataOriginalSize])
`
	case "ipv4":
		return `proc decode(): seq[byte] =
    result = newSeq[byte](dataOriginalSize)
    var offset = 0
    for addr in data:
        for part in addr.split('.'):
            if offset >= dataOriginalSize: return
            result[offset] = byte(parseInt(part))
            offset.inc
`
	case "ipv6", "mac":
		// both strip colons and parse hex pairs
		return `proc decode(): seq[byte] =
    result = newSeq[byte](dataOriginalSize)
    var offset = 0
    for addr in data:
        let cleaned = addr.replace(":", "")
        var i = 0
        while i + 1 < cleaned.len:
            if offset >= dataOriginalSize: return
            result[offset] = byte(parseHexInt(cleaned[i .. i + 1]))
            offset.inc
            i += 2
`
	default:
		return fmt.Sprintf("# decoder for %s not implemented\n", encoded.Name)
	}
}

func init() { Register(&NimEmitter{}) }
