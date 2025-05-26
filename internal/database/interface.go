package database

import "database/sql"

// DBInterface defines an interface for accessing the database
type DBInterface interface {
	DB() *sql.DB
	Close() error
}

// ensure that Database implements the DBInterface interface
var _ DBInterface = (*Database)(nil)
