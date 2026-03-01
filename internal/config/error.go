package config

import (
	"errors"
	"fmt"
)

var (
	ErrConfigNotFound   = errors.New("config file not found")
	ErrConfigRead       = errors.New("failed to read config file")
	ErrConfigParse      = errors.New("failed to parse config file")
	ErrConfigValidation = errors.New("config validation failed")
)

type ConfigError struct {
	Op   string
	Path string
	Err  error
}

func (e *ConfigError) Error() string {
	return e.Err.Error()
}

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %q: %s", e.Field, e.Msg)
}
