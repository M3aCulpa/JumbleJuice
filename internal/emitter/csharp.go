package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"strings"
)

type CSharpEmitter struct{}

func (cs *CSharpEmitter) Language() string { return "csharp" }

func (cs *CSharpEmitter) Emit(encoded encoder.Encoded, raw bool) (string, error) {
	if raw {
		return cs.data(encoded), nil
	}
	var sb strings.Builder
	sb.WriteString("// === imports ===\n")
	sb.WriteString(cs.imports(encoded))
	sb.WriteString("\n// === data ===\n")
	sb.WriteString(cs.data(encoded))
	sb.WriteString("\n// === decoder ===\n")
	sb.WriteString(cs.decoder(encoded))
	return sb.String(), nil
}

func (cs *CSharpEmitter) imports(encoded encoder.Encoded) string {
	var sb strings.Builder
	sb.WriteString("using System;\n")
	switch encoded.Name {
	case "ipv6", "mac":
		sb.WriteString("using System.Globalization;\n")
	}
	return sb.String()
}

func (cs *CSharpEmitter) data(encoded encoder.Encoded) string {
	var sb strings.Builder

	switch encoded.Name {
	case "hex", "dec":
		sb.WriteString("static byte[] data = new byte[] {\n")
		for i, chunk := range encoded.Chunks {
			if i > 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString("    ")
			sb.WriteString(chunk)
		}
		sb.WriteString("\n};\n")

	case "b64":
		sb.WriteString("static string data =\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(" +")
			} else {
				sb.WriteString(";")
			}
			sb.WriteString("\n")
		}

	default:
		// ipv4, ipv6, mac
		sb.WriteString("static string[] data = new string[] {\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("};\n")
	}

	sb.WriteString(fmt.Sprintf("static int dataOriginalSize = %d;\n", encoded.Size))
	return sb.String()
}

func (cs *CSharpEmitter) decoder(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "hex", "dec":
		return `static byte[] Decode() {
    int n = Math.Min(data.Length, dataOriginalSize);
    byte[] result = new byte[n];
    Array.Copy(data, result, n);
    return result;
}
`
	case "b64":
		return `static byte[] Decode() {
    byte[] decoded = Convert.FromBase64String(data);
    byte[] result = new byte[dataOriginalSize];
    Array.Copy(decoded, result, dataOriginalSize);
    return result;
}
`
	case "ipv4":
		return `static byte[] Decode() {
    byte[] result = new byte[dataOriginalSize];
    int offset = 0;
    foreach (string addr in data) {
        foreach (string part in addr.Split('.')) {
            if (offset >= dataOriginalSize) break;
            result[offset++] = (byte)int.Parse(part);
        }
    }
    return result;
}
`
	case "ipv6", "mac":
		// both strip colons and parse hex pairs
		return `static byte[] Decode() {
    byte[] result = new byte[dataOriginalSize];
    int offset = 0;
    foreach (string addr in data) {
        string cleaned = addr.Replace(":", "");
        for (int i = 0; i + 1 < cleaned.Length; i += 2) {
            if (offset >= dataOriginalSize) break;
            result[offset++] = byte.Parse(cleaned.Substring(i, 2), NumberStyles.HexNumber);
        }
    }
    return result;
}
`
	default:
		return fmt.Sprintf("// decoder for %s not implemented\n", encoded.Name)
	}
}

func init() { Register(&CSharpEmitter{}) }
