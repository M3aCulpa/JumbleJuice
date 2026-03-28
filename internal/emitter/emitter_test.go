package emitter

import (
	"jumblejuice/internal/encoder"
	"strings"
	"testing"
)

var allLanguages = []string{"c", "csharp", "go", "python", "rust", "nim"}
var allEncoders = []string{"hex", "b64", "dec", "ipv4", "ipv6", "mac"}

func ipv4Encoded() encoder.Encoded {
	return encoder.Encoded{
		Name:     "ipv4",
		Chunks:   []string{"72.101.108.108", "111.0.0.0"},
		Size:     5,
		Checksum: "abc123",
	}
}

func TestSnippetMarkers(t *testing.T) {
	enc := ipv4Encoded()
	for _, lang := range allLanguages {
		t.Run(lang, func(t *testing.T) {
			e, err := GetEmitter(lang)
			if err != nil {
				t.Fatalf("GetEmitter(%s): %v", lang, err)
			}
			code, err := e.Emit(enc, false)
			if err != nil {
				t.Fatalf("Emit(): %v", err)
			}
			for _, marker := range []string{"=== imports ===", "=== data ===", "=== decoder ==="} {
				if !strings.Contains(code, marker) {
					t.Errorf("missing marker %q", marker)
				}
			}
		})
	}
}

func TestRawMode(t *testing.T) {
	enc := ipv4Encoded()
	for _, lang := range allLanguages {
		t.Run(lang, func(t *testing.T) {
			e, _ := GetEmitter(lang)
			code, err := e.Emit(enc, true)
			if err != nil {
				t.Fatalf("Emit(): %v", err)
			}
			if strings.Contains(code, "=== imports ===") {
				t.Error("raw mode should not contain imports marker")
			}
			if strings.Contains(code, "=== decoder ===") {
				t.Error("raw mode should not contain decoder marker")
			}
			if code == "" {
				t.Error("raw output should not be empty")
			}
		})
	}
}

func TestAllCombinations(t *testing.T) {
	for _, encName := range allEncoders {
		enc, _ := encoder.GetEncoder(encName)
		encoded, _ := enc.Encode([]byte("Hello, World!"))
		for _, lang := range allLanguages {
			t.Run(encName+"+"+lang, func(t *testing.T) {
				e, _ := GetEmitter(lang)
				code, err := e.Emit(encoded, false)
				if err != nil {
					t.Fatalf("Emit(): %v", err)
				}
				if code == "" {
					t.Error("output should not be empty")
				}
			})
		}
	}
}

func TestEmptyInput(t *testing.T) {
	for _, encName := range allEncoders {
		enc, _ := encoder.GetEncoder(encName)
		encoded, _ := enc.Encode([]byte{})
		for _, lang := range allLanguages {
			t.Run(encName+"+"+lang, func(t *testing.T) {
				e, _ := GetEmitter(lang)
				code, err := e.Emit(encoded, false)
				if err != nil {
					t.Fatalf("Emit(): %v", err)
				}
				if !strings.Contains(code, "=== data ===") {
					t.Error("empty input should still produce data section")
				}
			})
		}
	}
}

func TestRegistry(t *testing.T) {
	registered := ListEmitters()
	for _, name := range allLanguages {
		found := false
		for _, r := range registered {
			if r == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("emitter %s not registered", name)
		}
	}
}
