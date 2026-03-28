# JumbleJuice

Binary payload encoder for red team operations. Takes a shellcode file, encodes it with one of several schemes, and outputs a code snippet you can paste into your payload source -> the encoded data array, a decoder function, and the required imports.

Supports C, C#, Go, Python, Rust, and Nim.

## Warning

This tool generates code that can execute arbitrary payloads. Use it only on systems you own or have explicit written authorization to test.

## Build

```
go build -o jj cmd/jj/main.go
```

Requires Go 1.22+. No external dependencies.

## Usage

```
jj -i <input> [-e encoder] [-t language] [-o file] [--raw]
```

Output goes to stdout by default. Pipe it, redirect it, or copy it.

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | (required) | Input file (use `-` for stdin) |
| `-e` | `ipv4` | Encoder: `hex`, `b64`, `dec`, `ipv4`, `ipv6`, `mac` |
| `-t` | `c` | Language: `c`, `csharp`, `go`, `python`, `rust`, `nim` |
| `-o` | stdout | Write to file instead of stdout |
| `--raw` | | Data blob only -- no decoder, no imports |

## Examples

Encode shellcode as IPv4 addresses, get a C snippet:
```
jj -i shellcode.bin -e ipv4 -t c
```

Get a Rust snippet using hex encoding:
```
jj -i shellcode.bin -e hex -t rust
```

Just the data array for C#, no decoder:
```
jj -i shellcode.bin -e mac -t csharp --raw
```

Save to a file:
```
jj -i shellcode.bin -e b64 -t python -o snippet.py
```

## Output format

The output has three labeled sections you can copy independently:

```
// === imports ===
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// === data ===
static const char* data[] = {
    "72.101.108.108",
    "111.0.0.0"
};
static const size_t data_count = 2;
static const size_t data_original_size = 5;

// === decoder ===
size_t decode(unsigned char* dst, size_t dst_len) {
    ...
}
```

The section markers use the target language's comment syntax (`//` for C/C#/Go/Rust, `#` for Python/Nim).

## Encoders

| Name | Bytes per unit | Output |
|------|---------------|--------|
| `hex` | 1 | `0x48, 0x65, 0x6C, ...` |
| `b64` | 3 -> 4 chars | Standard base64 |
| `dec` | 1 | `72, 101, 108, ...` |
| `ipv4` | 4 | `72.101.108.108` |
| `ipv6` | 16 | Full IPv6 addresses |
| `mac` | 6 | `48:65:6C:6C:6F:21` |

Encoders that group multiple bytes (ipv4, ipv6, mac) pad the last chunk with zeros. The `data_original_size` constant in the output tells the decoder where to truncate.

## Project layout

```
cmd/jj/              CLI
internal/encoder/    Encoders (hex, b64, dec, ipv4, ipv6, mac)
internal/emitter/    Snippet generators (c, csharp, go, python, rust, nim)
internal/common/     File I/O, hashing
```

## Tests

```
go test -race ./...
```

## License

MIT -- see [LICENSE](LICENSE).
