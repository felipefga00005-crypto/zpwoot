package uuid

import (
	"github.com/google/uuid"
)

// Generator provides UUID generation functionality
type Generator struct{}

// New creates a new UUID generator
func New() *Generator {
	return &Generator{}
}

// Generate generates a new UUID v4
func (g *Generator) Generate() string {
	return uuid.New().String()
}

// GenerateShort generates a short UUID (first 8 characters)
func (g *Generator) GenerateShort() string {
	return uuid.New().String()[:8]
}

// Parse parses a UUID string
func (g *Generator) Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValid checks if a string is a valid UUID
func (g *Generator) IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// Global generator instance
var defaultGenerator = New()

// Generate generates a new UUID v4 using the default generator
func Generate() string {
	return defaultGenerator.Generate()
}

// GenerateShort generates a short UUID using the default generator
func GenerateShort() string {
	return defaultGenerator.GenerateShort()
}

// Parse parses a UUID string using the default generator
func Parse(s string) (uuid.UUID, error) {
	return defaultGenerator.Parse(s)
}

// IsValid checks if a string is a valid UUID using the default generator
func IsValid(s string) bool {
	return defaultGenerator.IsValid(s)
}
