package id

import (
	"crypto/rand"
	"encoding/binary"
	"strings"
)

const (
	// Base62Chars contains all characters used in base62 encoding
	Base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// KeyLength is the length of generated keys
	KeyLength = 8
)

// Generator handles the generation of unique IDs
type Generator struct {
	chars string
}

// NewGenerator creates a new ID generator
func NewGenerator() *Generator {
	return &Generator{
		chars: Base62Chars,
	}
}

// Generate creates a new random base62 encoded ID
func (g *Generator) Generate() (string, error) {
	// Generate 48 bits (6 bytes) of random data
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	// Convert to uint64 for easier manipulation
	num := binary.BigEndian.Uint64(append(make([]byte, 2), buf...))

	// Convert to base62
	var builder strings.Builder
	builder.Grow(KeyLength)

	// Fill the key to exact length
	for i := 0; i < KeyLength; i++ {
		builder.WriteByte(g.chars[num%62])
		num /= 62
	}

	return builder.String(), nil
}

// ValidateKey checks if a key matches our requirements
func (g *Generator) ValidateKey(key string) bool {
	if len(key) != KeyLength {
		return false
	}

	for _, c := range key {
		if !strings.ContainsRune(g.chars, c) {
			return false
		}
	}

	return true
}
