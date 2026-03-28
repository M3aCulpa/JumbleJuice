package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"strings"
)

type CEmitter struct{}

func (c *CEmitter) Language() string { return "c" }

func (c *CEmitter) Emit(encoded encoder.Encoded, raw bool) (string, error) {
	if raw {
		return c.data(encoded), nil
	}
	var sb strings.Builder
	sb.WriteString("// === imports ===\n")
	sb.WriteString(c.imports(encoded))
	sb.WriteString("\n// === data ===\n")
	sb.WriteString(c.data(encoded))
	sb.WriteString("\n// === decoder ===\n")
	sb.WriteString(c.decoder(encoded))
	return sb.String(), nil
}

func (c *CEmitter) imports(encoded encoder.Encoded) string {
	var sb strings.Builder
	sb.WriteString("#include <stdlib.h>\n")
	sb.WriteString("#include <string.h>\n")
	sb.WriteString("#include <stdint.h>\n")
	switch encoded.Name {
	case "ipv4", "ipv6", "mac":
		sb.WriteString("#include <stdio.h>\n")
	}
	return sb.String()
}

func (c *CEmitter) data(encoded encoder.Encoded) string {
	var sb strings.Builder

	switch encoded.Name {
	case "hex", "dec":
		// both produce unsigned char arrays from comma-separated chunks
		sb.WriteString("static const unsigned char data[] = {\n")
		for i, chunk := range encoded.Chunks {
			if i > 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString("    ")
			sb.WriteString(chunk)
		}
		sb.WriteString("\n};\n")

	default:
		// b64, ipv4, ipv6, mac all produce string arrays
		sb.WriteString("static const char* data[] = {\n")
		for i, chunk := range encoded.Chunks {
			sb.WriteString(fmt.Sprintf("    \"%s\"", chunk))
			if i < len(encoded.Chunks)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("};\n")
	}

	sb.WriteString(fmt.Sprintf("static const size_t data_count = %d;\n", len(encoded.Chunks)))
	sb.WriteString(fmt.Sprintf("static const size_t data_original_size = %d;\n", encoded.Size))
	return sb.String()
}

func (c *CEmitter) decoder(encoded encoder.Encoded) string {
	switch encoded.Name {
	case "hex", "dec":
		return `size_t decode(unsigned char* dst, size_t dst_len) {
    size_t n = sizeof(data);
    if (n > dst_len) n = dst_len;
    if (n > data_original_size) n = data_original_size;
    memcpy(dst, data, n);
    return n;
}
`
	case "b64":
		return `static const unsigned char b64_lut[256] = {
    ['A']=0,  ['B']=1,  ['C']=2,  ['D']=3,  ['E']=4,  ['F']=5,
    ['G']=6,  ['H']=7,  ['I']=8,  ['J']=9,  ['K']=10, ['L']=11,
    ['M']=12, ['N']=13, ['O']=14, ['P']=15, ['Q']=16, ['R']=17,
    ['S']=18, ['T']=19, ['U']=20, ['V']=21, ['W']=22, ['X']=23,
    ['Y']=24, ['Z']=25, ['a']=26, ['b']=27, ['c']=28, ['d']=29,
    ['e']=30, ['f']=31, ['g']=32, ['h']=33, ['i']=34, ['j']=35,
    ['k']=36, ['l']=37, ['m']=38, ['n']=39, ['o']=40, ['p']=41,
    ['q']=42, ['r']=43, ['s']=44, ['t']=45, ['u']=46, ['v']=47,
    ['w']=48, ['x']=49, ['y']=50, ['z']=51, ['0']=52, ['1']=53,
    ['2']=54, ['3']=55, ['4']=56, ['5']=57, ['6']=58, ['7']=59,
    ['8']=60, ['9']=61, ['+']=62, ['/']=63
};

size_t decode(unsigned char* dst, size_t dst_len) {
    size_t out = 0;
    for (size_t i = 0; i < data_count; i++) {
        const char* chunk = data[i];
        size_t len = strlen(chunk);
        for (size_t j = 0; j < len; j += 4) {
            uint32_t val = 0;
            int bits = 0;
            for (int k = 0; k < 4 && j + k < len; k++) {
                unsigned char ch = (unsigned char)chunk[j + k];
                if (ch == '=') break;
                val = (val << 6) | b64_lut[ch];
                bits += 6;
            }
            while (bits >= 8) {
                bits -= 8;
                if (out >= dst_len || out >= data_original_size) return out;
                dst[out++] = (unsigned char)((val >> bits) & 0xFF);
            }
        }
    }
    return out;
}
`
	case "ipv4":
		return `size_t decode(unsigned char* dst, size_t dst_len) {
    size_t out = 0;
    for (size_t i = 0; i < data_count; i++) {
        unsigned int a, b, c, d;
        if (sscanf(data[i], "%u.%u.%u.%u", &a, &b, &c, &d) != 4) continue;
        if (out < dst_len && out < data_original_size) dst[out++] = (unsigned char)a;
        if (out < dst_len && out < data_original_size) dst[out++] = (unsigned char)b;
        if (out < dst_len && out < data_original_size) dst[out++] = (unsigned char)c;
        if (out < dst_len && out < data_original_size) dst[out++] = (unsigned char)d;
    }
    return out;
}
`
	case "ipv6", "mac":
		// both strip colons and parse hex pairs
		return `static unsigned char hex_val(char ch) {
    if (ch >= '0' && ch <= '9') return (unsigned char)(ch - '0');
    if (ch >= 'a' && ch <= 'f') return (unsigned char)(ch - 'a' + 10);
    if (ch >= 'A' && ch <= 'F') return (unsigned char)(ch - 'A' + 10);
    return 0;
}

size_t decode(unsigned char* dst, size_t dst_len) {
    size_t out = 0;
    for (size_t i = 0; i < data_count; i++) {
        const char* s = data[i];
        char hex[33];
        int h = 0;
        while (*s && h < 32) {
            if (*s != ':') hex[h++] = *s;
            s++;
        }
        hex[h] = '\0';
        for (int j = 0; j + 1 < h; j += 2) {
            if (out >= dst_len || out >= data_original_size) return out;
            dst[out++] = (hex_val(hex[j]) << 4) | hex_val(hex[j + 1]);
        }
    }
    return out;
}
`

	default:
		return fmt.Sprintf("/* decoder for %s not implemented */\n", encoded.Name)
	}
}

func init() { Register(&CEmitter{}) }
