package logger

import (
	"time"
)

// LogConfig represents logger configuration
type LogConfig struct {
	Level      string          `json:"level" yaml:"level" env:"LOG_LEVEL"`
	Format     string          `json:"format" yaml:"format" env:"LOG_FORMAT"`
	Output     string          `json:"output" yaml:"output" env:"LOG_OUTPUT"`
	TimeFormat string          `json:"timeFormat" yaml:"timeFormat" env:"LOG_TIME_FORMAT"`
	Caller     bool            `json:"caller" yaml:"caller" env:"LOG_CALLER"`
	Sampling   *SamplingConfig `json:"sampling,omitempty" yaml:"sampling,omitempty"`
}

// SamplingConfig represents sampling configuration for high-volume logging
type SamplingConfig struct {
	Initial    int           `json:"initial" yaml:"initial"`
	Thereafter int           `json:"thereafter" yaml:"thereafter"`
	Tick       time.Duration `json:"tick" yaml:"tick"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		TimeFormat: time.RFC3339,
		Caller:     false,
	}
}

// DevelopmentConfig returns development-friendly logger configuration
func DevelopmentConfig() *LogConfig {
	return &LogConfig{
		Level:      "debug",
		Format:     "console",
		Output:     "stdout",
		TimeFormat: "2006-01-02 15:04:05",
		Caller:     true,
	}
}

// ProductionConfig returns production-optimized logger configuration
func ProductionConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		TimeFormat: time.RFC3339,
		Caller:     false,
		Sampling: &SamplingConfig{
			Initial:    100,
			Thereafter: 100,
			Tick:       time.Second,
		},
	}
}

// Validate validates the logger configuration
func (c *LogConfig) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"trace": true,
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}

	if !validLevels[c.Level] {
		c.Level = "info" // Default fallback
	}

	// Validate format
	validFormats := map[string]bool{
		"json":    true,
		"console": true,
	}

	if !validFormats[c.Format] {
		c.Format = "json" // Default fallback
	}

	// Validate output
	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
	}

	if !validOutputs[c.Output] {
		c.Output = "stdout" // Default fallback
	}

	// Set default time format if empty
	if c.TimeFormat == "" {
		c.TimeFormat = time.RFC3339
	}

	return nil
}

// IsDebugEnabled returns true if debug level is enabled
func (c *LogConfig) IsDebugEnabled() bool {
	return c.Level == "debug" || c.Level == "trace"
}

// IsProductionMode returns true if this is a production configuration
func (c *LogConfig) IsProductionMode() bool {
	return c.Format == "json" && c.Level != "debug" && c.Level != "trace"
}
