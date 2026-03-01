package config

import (
	"fmt"
	"slices"
)

type LogConfig struct {
	// Variants:
	//  - "debug"
	//  - "info"
	//  - "warning"
	//  - "error"
	Level string `yaml:"level"`

	// Variants:
	//  - "stdout"
	//  - "file"
	//  - "both"
	Output string `yaml:"output"`

	// Variants:
	//  - "text"
	//  - "json"
	Format    string `yaml:"format"`
	Path      string `yaml:"path"`
	AddSource bool   `yaml:"add_source"`
}

func (c *LogConfig) Validate() error {
	validLevels := []string{"debug", "info", "warning", "error"}
	validOutputs := []string{"stdout", "file", "both"}
	validFormats := []string{"text", "json"}

	if !slices.Contains(validLevels, c.Level) {
		return &ValidationError{Field: "log.level", Msg: fmt.Sprintf("unknown value %q", c.Level)}
	}

	if !slices.Contains(validOutputs, c.Output) {
		return &ValidationError{Field: "log.output", Msg: fmt.Sprintf("unknown value %q", c.Output)}
	}

	if !slices.Contains(validFormats, c.Format) {
		return &ValidationError{Field: "log.format", Msg: fmt.Sprintf("unknown value %q", c.Format)}
	}

	if (c.Output == "file" || c.Output == "both") && c.Path == "" {
		return &ValidationError{Field: "log.path", Msg: "path is required when output is \"file\" or \"both\""}
	}

	return nil
}
