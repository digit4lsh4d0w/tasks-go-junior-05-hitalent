package config

import (
	"fmt"
	"slices"
)

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

func (c *DBConfig) Validate() error {
	validDrivers := []string{"sqlite", "postgres"}

	if !slices.Contains(validDrivers, c.Driver) {
		return &ValidationError{Field: "database.driver", Msg: fmt.Sprintf("unknown driver %q", c.Driver)}
	}

	if c.DSN == "" {
		return &ValidationError{Field: "database.dsn", Msg: "dsn is required"}
	}

	return nil
}
