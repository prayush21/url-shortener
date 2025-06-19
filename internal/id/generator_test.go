package id

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	assert.NotNil(t, g)
	assert.Equal(t, Base62Chars, g.chars)
}

func TestGenerator_Generate(t *testing.T) {
	g := NewGenerator()

	// Generate multiple keys and verify their properties
	for i := 0; i < 100; i++ {
		key, err := g.Generate()
		assert.NoError(t, err)
		assert.Len(t, key, KeyLength)
		assert.True(t, g.ValidateKey(key))
	}

	// Verify uniqueness of generated keys
	keys := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		key, err := g.Generate()
		assert.NoError(t, err)
		assert.False(t, keys[key], "Duplicate key generated: %s", key)
		keys[key] = true
	}
}

func TestGenerator_ValidateKey(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "Valid key",
			key:   "aB1cD2eF",
			valid: true,
		},
		{
			name:  "Too short",
			key:   "abc123",
			valid: false,
		},
		{
			name:  "Too long",
			key:   "abc123def456",
			valid: false,
		},
		{
			name:  "Invalid characters",
			key:   "abc!@#$%^",
			valid: false,
		},
		{
			name:  "Empty string",
			key:   "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, g.ValidateKey(tt.key))
		})
	}
}

func TestGenerator_Generate_Distribution(t *testing.T) {
	g := NewGenerator()
	charCount := make(map[rune]int)
	totalChars := 0

	// Generate a large number of keys to check distribution
	for i := 0; i < 10000; i++ {
		key, err := g.Generate()
		assert.NoError(t, err)

		for _, c := range key {
			charCount[c]++
			totalChars++
		}
	}

	// Check that all possible characters are used
	for _, c := range Base62Chars {
		count := charCount[c]
		assert.Greater(t, count, 0, "Character %c was never used", c)
	}

	// Check that no unexpected characters are used
	assert.Equal(t, len(Base62Chars), len(charCount), "Unexpected characters in generated keys")
}

func TestGenerator_Generate_RandomError(t *testing.T) {
	// Replace rand.Reader with a reader that always fails
	originalReader := rand.Reader
	rand.Reader = failingReader{}
	defer func() {
		rand.Reader = originalReader
	}()

	g := NewGenerator()
	_, err := g.Generate()
	assert.Error(t, err)
}

// failingReader is an io.Reader that always returns an error
type failingReader struct{}

func (failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
