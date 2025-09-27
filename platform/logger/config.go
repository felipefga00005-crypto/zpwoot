package logger

// LogConfig represents zpwoot logger configuration
type LogConfig struct {
	Level  string `json:"level" yaml:"level" env:"LOG_LEVEL"`
	Format string `json:"format" yaml:"format" env:"LOG_FORMAT"`
	Output string `json:"output" yaml:"output" env:"LOG_OUTPUT"`
	Caller bool   `json:"caller" yaml:"caller" env:"LOG_CALLER"`
}

// DevelopmentConfig returns development-friendly logger configuration
// Features: ConsoleWriter, debug level, caller info, colorized output
func DevelopmentConfig() *LogConfig {
	return &LogConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
		Caller: true,
	}
}

// ProductionConfig returns production-optimized logger configuration
// Features: JSON format, info level, no caller info, structured logging
func ProductionConfig() *LogConfig {
	return &LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
		Caller: false,
	}
}

// Validate validates and sets defaults for logger configuration
func (c *LogConfig) Validate() {
	// Validate and set default log level
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[c.Level] {
		c.Level = "info"
	}

	// Validate and set default format
	if c.Format != "console" && c.Format != "json" {
		c.Format = "json"
	}

	// Validate and set default output
	if c.Output != "stdout" && c.Output != "stderr" && c.Output != "file" {
		c.Output = "stdout"
	}
}

// IsDevelopment returns true if this is a development configuration
func (c *LogConfig) IsDevelopment() bool {
	return c.Format == "console" && (c.Level == "debug" || c.Level == "trace")
}

// IsProduction returns true if this is a production configuration
func (c *LogConfig) IsProduction() bool {
	return c.Format == "json" && c.Level != "debug" && c.Level != "trace"
}
