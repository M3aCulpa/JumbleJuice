package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"strings"
)

type RustEmitter struct{}

func (r *RustEmitter) Language() string { return "rust" }

func (r *RustEmitter) Emit(encoded encoder.Encoded, raw bool) (string, error) {
	if raw {
		return r.data(encoded), nil
	}
	var sb strings.Builder
	sb.WriteString("// === imports ===\n")
	sb.WriteString(r.imports())
	sb.WriteString("\n// === data ===\n")
	sb.WriteString(r.data(encoded))
	sb.WriteString("\n// === decoder ===\n")
	sb.WriteString(r.decoder(encoded))
	return sb.String(), nil
}

func (r *RustEmitter) imports() string {
	return "// (no imports required)\n"
}

func (r *RustEmitter) data(encoded encoder.Encoded) string {
	var sb strings.Builder

	switch encoded.Name {
	case "hex", "dec":
		sb.WriteString("const DATA: &[u8] = &[\n")
		for i, chunk := range encoded.Chunks {
			if i > 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString("    ")
			sb.WriteString(chunk)
		}
		sb.WriteString("\n];\n")

	default:
		// b64, ipv4, ipv6, mac all produce string slices
		sb.WriteString("const DATA: &[&str] = &[\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("];\n")
	}

	sb.WriteString(fmt.Sprintf("const DATA_ORIGINAL_SIZE: usize = %d;\n", encoded.Size))
	return sb.String()
}

func (r *RustEmitter) decoder(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "hex", "dec":
		return `fn decode() -> Vec<u8> {
    DATA[..DATA_ORIGINAL_SIZE].to_vec()
}
`
	case "b64":
		return `fn decode() -> Vec<u8> {
    const LUT: [u8; 256] = {
        let mut table = [255u8; 256];
        let alphabet = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
        let mut i = 0;
        while i < 64 {
            table[alphabet[i] as usize] = i as u8;
            i += 1;
        }
        table
    };
    let mut result = Vec::with_capacity(DATA_ORIGINAL_SIZE);
    for chunk in DATA {
        let bytes = chunk.as_bytes();
        let mut i = 0;
        while i < bytes.len() {
            let mut val: u32 = 0;
            let mut bits: u32 = 0;
            for k in 0..4 {
                if i + k >= bytes.len() { break; }
                let ch = bytes[i + k];
                if ch == b'=' { break; }
                val = (val << 6) | LUT[ch as usize] as u32;
                bits += 6;
            }
            i += 4;
            while bits >= 8 {
                bits -= 8;
                if result.len() >= DATA_ORIGINAL_SIZE { return result; }
                result.push(((val >> bits) & 0xFF) as u8);
            }
        }
    }
    result
}
`
	case "ipv4":
		return `fn decode() -> Vec<u8> {
    let mut result = Vec::with_capacity(DATA_ORIGINAL_SIZE);
    for addr in DATA {
        for part in addr.split('.') {
            if result.len() >= DATA_ORIGINAL_SIZE { return result; }
            if let Ok(b) = part.parse::<u8>() {
                result.push(b);
            }
        }
    }
    result
}
`
	case "ipv6", "mac":
		// both strip colons and parse hex pairs
		return `fn decode() -> Vec<u8> {
    let mut result = Vec::with_capacity(DATA_ORIGINAL_SIZE);
    for addr in DATA {
        let cleaned: String = addr.chars().filter(|c| *c != ':').collect();
        let bytes = cleaned.as_bytes();
        let mut i = 0;
        while i + 1 < bytes.len() {
            if result.len() >= DATA_ORIGINAL_SIZE { return result; }
            let s = &cleaned[i..i + 2];
            if let Ok(b) = u8::from_str_radix(s, 16) {
                result.push(b);
            }
            i += 2;
        }
    }
    result
}
`
	default:
		return fmt.Sprintf("// decoder for %s not implemented\n", encoded.Name)
	}
}

func init() { Register(&RustEmitter{}) }
