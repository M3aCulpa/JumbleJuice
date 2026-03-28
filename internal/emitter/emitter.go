package emitter

import (
	"fmt"
	"jumblejuice/internal/encoder"
	"sync"
)

// generates code snippets for a target language.
type Emitter interface {
	Language() string
	Emit(encoded encoder.Encoded, raw bool) (string, error)
}

var (
	emitters = make(map[string]Emitter)
	mu       sync.RWMutex
)

func Register(emitter Emitter) {
	mu.Lock()
	defer mu.Unlock()
	emitters[emitter.Language()] = emitter
}

func GetEmitter(language string) (Emitter, error) {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := emitters[language]
	if !ok {
		return nil, fmt.Errorf("emitter not found for language: %s", language)
	}
	return e, nil
}

func ListEmitters() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(emitters))
	for name := range emitters {
		names = append(names, name)
	}
	return names
}
