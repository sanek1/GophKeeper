package database

import "database/sql"

// DBInterface определяет интерфейс для доступа к базе данных
type DBInterface interface {
	DB() *sql.DB
	Close() error
}

// Убеждаемся, что Database реализует интерфейс DBInterface
var _ DBInterface = (*Database)(nil)
