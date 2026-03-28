package encoder

import (
	"bytes"
	"testing"
)

// round-trip tests for all encoders
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		enc  Encoder
		data []byte
	}{
		{"hex/hello", &HexEncoder{}, []byte("Hello, World!")},
		{"hex/binary", &HexEncoder{}, []byte{0, 1, 127, 128, 255}},
		{"b64/hello", &Base64Encoder{}, []byte("Hello, World!")},
		{"b64/binary", &Base64Encoder{}, []byte{0, 1, 127, 128, 255}},
		{"dec/hello", &DecimalEncoder{}, []byte("Hello, World!")},
		{"dec/single", &DecimalEncoder{}, []byte{72}},
		{"ipv4/exact", &IPv4Encoder{}, []byte{10, 0, 0, 1, 192, 168, 0, 1}},
		{"ipv4/padding", &IPv4Encoder{}, []byte("Hello")},
		{"ipv6/exact16", &IPv6Encoder{}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
		{"ipv6/padding", &IPv6Encoder{}, []byte("Hello, World!")},
		{"ipv6/zeros", &IPv6Encoder{}, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{"mac/exact6", &MACEncoder{}, []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}},
		{"mac/padding", &MACEncoder{}, []byte("Hi")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := tt.enc.Encode(tt.data)
			if err != nil {
				t.Fatalf("Encode(): %v", err)
			}
			decoded, err := tt.enc.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode(): %v", err)
			}
			if !bytes.Equal(decoded, tt.data) {
				t.Errorf("round-trip failed: got %v, want %v", decoded, tt.data)
			}
		})
	}
}

func TestEmptyInput(t *testing.T) {
	encoders := []Encoder{
		&HexEncoder{}, &Base64Encoder{}, &DecimalEncoder{},
		&IPv4Encoder{}, &IPv6Encoder{}, &MACEncoder{},
	}
	for _, enc := range encoders {
		t.Run(enc.Name(), func(t *testing.T) {
			encoded, err := enc.Encode([]byte{})
			if err != nil {
				t.Fatalf("Encode(): %v", err)
			}
			if len(encoded.Chunks) != 0 {
				t.Errorf("empty input should produce 0 chunks, got %d", len(encoded.Chunks))
			}
		})
	}
}

func TestIPv4ChunkValues(t *testing.T) {
	enc := &IPv4Encoder{}
	encoded, _ := enc.Encode([]byte("Hello"))

	expected := []string{"72.101.108.108", "111.0.0.0"}
	if len(encoded.Chunks) != len(expected) {
		t.Fatalf("chunks = %d, want %d", len(encoded.Chunks), len(expected))
	}
	for i, want := range expected {
		if encoded.Chunks[i] != want {
			t.Errorf("chunk[%d] = %s, want %s", i, encoded.Chunks[i], want)
		}
	}
}

func TestMACChunkValues(t *testing.T) {
	enc := &MACEncoder{}
	encoded, _ := enc.Encode([]byte("Hello!"))

	if len(encoded.Chunks) != 1 || encoded.Chunks[0] != "48:65:6C:6C:6F:21" {
		t.Errorf("unexpected chunks: %v", encoded.Chunks)
	}
}

func TestIPv6Compression(t *testing.T) {
	tests := []struct{ input, want string }{
		{"1:2:3:4:5:6:7:8", "1:2:3:4:5:6:7:8"},
		{"1:0:0:0:0:0:0:8", "1::8"},
		{"0:0:0:0:0:0:0:0", "::"},
		{"1:0:0:4:0:0:0:8", "1:0:0:4::8"},
		{"0:0:0:0:0:0:0:1", "::1"},
	}
	for _, tt := range tests {
		if got := compressIPv6(tt.input); got != tt.want {
			t.Errorf("compressIPv6(%s) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestIPv6Expansion(t *testing.T) {
	tests := []struct{ input, want string }{
		{"1:2:3:4:5:6:7:8", "1:2:3:4:5:6:7:8"},
		{"1::8", "1:0:0:0:0:0:0:8"},
		{"::", "0:0:0:0:0:0:0:0"},
		{"::1", "0:0:0:0:0:0:0:1"},
		{"1::", "1:0:0:0:0:0:0:0"},
	}
	for _, tt := range tests {
		if got := expandIPv6(tt.input); got != tt.want {
			t.Errorf("expandIPv6(%s) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestRegistry(t *testing.T) {
	expected := []string{"hex", "b64", "dec", "ipv4", "ipv6", "mac"}
	registered := ListEncoders()

	for _, name := range expected {
		found := false
		for _, r := range registered {
			if r == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("encoder %s not registered", name)
		}
	}

	if _, err := GetEncoder("nonexistent"); err == nil {
		t.Error("GetEncoder(nonexistent) should return error")
	}
}
