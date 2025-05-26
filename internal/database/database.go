package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/yourusername/gophkeeper/internal/config"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s port=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode, cfg.DBPort,
	)
	fmt.Println(connStr)

	db, err := sql.Open("postgres", "postgres://postgres:admin@localhost:5432/gophkeeper?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) DB() *sql.DB {
	return d.db
}
