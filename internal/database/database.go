package database

import (
	"fmt"
	"hitalent-task/internal/config"
	"hitalent-task/internal/errors"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.Driver == "" {
		return nil, &errors.DatabaseError{DSN: cfg.DSN, Err: fmt.Errorf("driver is required")}
	}

	if cfg.DSN == "" {
		return nil, &errors.DatabaseError{DSN: cfg.DSN, Err: fmt.Errorf("DSN is required")}
	}

	var db *gorm.DB
	var err error

	switch cfg.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, &errors.DatabaseError{DSN: cfg.DSN, Err: gorm.ErrUnsupportedDriver}
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}
