package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	originalEnvs := map[string]string{
		"SERVER_PORT": os.Getenv("SERVER_PORT"),
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASS":     os.Getenv("DB_PASS"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_SSL_MODE": os.Getenv("DB_SSL_MODE"),
		"JWT_SECRET":  os.Getenv("JWT_SECRET"),
	}

	defer func() {
		for key, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("DefaultConfig", func(t *testing.T) {
		for key := range originalEnvs {
			os.Unsetenv(key)
		}

		cfg, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "8081", cfg.ServerPort)
		assert.Equal(t, "localhost", cfg.DBHost)
		assert.Equal(t, "5432", cfg.DBPort)
		assert.Equal(t, "admin", cfg.DBUser)
		assert.Equal(t, "admin", cfg.DBPass)
		assert.Equal(t, "gophkeeper", cfg.DBName)
		assert.Equal(t, "disable", cfg.DBSSLMode)
		assert.Equal(t, DefaultJWTSecret, cfg.JWTSecret)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "9000")
		os.Setenv("DB_HOST", "test.host")
		os.Setenv("DB_PORT", "3306")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASS", "testpass")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSL_MODE", "require")
		os.Setenv("JWT_SECRET", "test-secret")

		cfg, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "9000", cfg.ServerPort)
		assert.Equal(t, "test.host", cfg.DBHost)
		assert.Equal(t, "3306", cfg.DBPort)
		assert.Equal(t, "testuser", cfg.DBUser)
		assert.Equal(t, "testpass", cfg.DBPass)
		assert.Equal(t, "testdb", cfg.DBName)
		assert.Equal(t, "require", cfg.DBSSLMode)
		assert.Equal(t, "test-secret", cfg.JWTSecret)
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("InvalidServerPort", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "invalid",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "admin",
			DBName:     "gophkeeper",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid server port")
	})

	t.Run("InvalidDBPort", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "invalid",
			DBUser:     "admin",
			DBName:     "gophkeeper",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid database port")
	})

	t.Run("EmptyDBHost", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "",
			DBPort:     "5432",
			DBUser:     "admin",
			DBName:     "gophkeeper",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database host is required")
	})

	t.Run("EmptyDBUser", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "",
			DBName:     "gophkeeper",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database user is required")
	})

	t.Run("EmptyDBName", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "admin",
			DBName:     "",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database name is required")
	})

	t.Run("EmptyJWTSecret", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "admin",
			DBName:     "gophkeeper",
			JWTSecret:  "",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret is required")
	})

	t.Run("InvalidSSLMode", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "admin",
			DBName:     "gophkeeper",
			DBSSLMode:  "invalid",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid SSL mode")
	})

	t.Run("ValidConfig", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "admin",
			DBName:     "gophkeeper",
			DBSSLMode:  "disable",
			JWTSecret:  "secret",
		}
		err := cfg.validate()
		assert.NoError(t, err)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("EnvVarExists", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnv("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("EnvVarNotExists", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_VAR")

		result := getEnv("NONEXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})
}
