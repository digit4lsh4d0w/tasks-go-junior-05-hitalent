package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBConfig  DBConfig  `yaml:"database"`
	LogConfig LogConfig `yaml:"log"`
}

func (c *Config) Validate() error {
	validators := []interface{ Validate() error }{
		&c.DBConfig,
		&c.LogConfig,
	}

	for _, v := range validators {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrConfigValidation, err)
		}
	}

	return nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, path)
		}
		return nil, &ConfigError{Op: "read", Path: path, Err: err}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, &ConfigError{Op: "parse", Path: path, Err: err}
	}

	if err := cfg.Validate(); err != nil {
		return nil, &ConfigError{Op: "validate", Path: path, Err: err}
	}

	return &cfg, nil
}
