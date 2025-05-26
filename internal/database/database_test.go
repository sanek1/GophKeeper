package database

import (
	"testing"

	"github.com/sanek1/GophKeeper/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		cfg := &config.Config{
			DBHost:    "localhost",
			DBPort:    "5432",
			DBUser:    "postgres",
			DBPass:    "admin",
			DBName:    "gophkeeper",
			DBSSLMode: "disable",
		}

		// Тест создания объекта database с валидной конфигурацией
		// Примечание: этот тест может упасть если PostgreSQL не запущен
		db, err := NewDatabase(cfg)
		if err != nil {
			// Если база данных недоступна, проверяем, что ошибка корректная
			assert.Contains(t, err.Error(), "failed to")
		} else {
			// Если подключение успешно, проверяем что database создан
			assert.NotNil(t, db)
			assert.NotNil(t, db.DB())
			err = db.Close()
			assert.NoError(t, err)
		}
	})

	t.Run("InvalidDriver", func(t *testing.T) {
		cfg := &config.Config{
			DBHost:    "invalid://host",
			DBPort:    "invalid",
			DBUser:    "invalid",
			DBPass:    "invalid",
			DBName:    "invalid",
			DBSSLMode: "invalid",
		}

		// Тест с невалидной конфигурацией
		_, err := NewDatabase(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to")
	})
}

func TestDatabase_DB(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		DBHost:    "localhost",
		DBPort:    "5432",
		DBUser:    "postgres",
		DBPass:    "admin",
		DBName:    "gophkeeper",
		DBSSLMode: "disable",
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		// Если база данных недоступна, пропускаем тест
		t.Skip("Database not available:", err)
		return
	}
	defer db.Close()

	// Проверяем, что метод DB() возвращает корректный объект
	sqlDB := db.DB()
	assert.NotNil(t, sqlDB)
}

func TestDatabase_Close(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		DBHost:    "localhost",
		DBPort:    "5432",
		DBUser:    "postgres",
		DBPass:    "admin",
		DBName:    "gophkeeper",
		DBSSLMode: "disable",
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		// Если база данных недоступна, пропускаем тест
		t.Skip("Database not available:", err)
		return
	}

	// Проверяем, что Close() работает без ошибок
	err = db.Close()
	assert.NoError(t, err)
}
