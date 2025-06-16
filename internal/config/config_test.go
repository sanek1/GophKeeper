package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		// Set required JWT_SECRET for tests
		os.Setenv("JWT_SECRET", "test-jwt-secret-key-that-is-long-enough-for-validation")
		defer os.Unsetenv("JWT_SECRET")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.ServerPort)
		assert.Equal(t, "localhost", cfg.DBHost)
		assert.Equal(t, "5432", cfg.DBPort)
		assert.Equal(t, "postgres", cfg.DBUser)
		assert.Equal(t, "admin", cfg.DBPass)
		assert.Equal(t, "gophkeeper", cfg.DBName)
		assert.Equal(t, "disable", cfg.DBSSLMode)
		assert.Equal(t, "test-jwt-secret-key-that-is-long-enough-for-validation", cfg.JWTSecret)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		// Set custom environment variables
		os.Setenv("SERVER_PORT", "9090")
		os.Setenv("DB_HOST", "custom-host")
		os.Setenv("DB_PORT", "3306")
		os.Setenv("DB_USER", "custom-user")
		os.Setenv("DB_PASS", "custom-pass")
		os.Setenv("DB_NAME", "custom-db")
		os.Setenv("DB_SSL_MODE", "require")
		os.Setenv("JWT_SECRET", "custom-jwt-secret-key-that-is-long-enough-for-security-validation")

		defer func() {
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASS")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("DB_SSL_MODE")
			os.Unsetenv("JWT_SECRET")
		}()

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "9090", cfg.ServerPort)
		assert.Equal(t, "custom-host", cfg.DBHost)
		assert.Equal(t, "3306", cfg.DBPort)
		assert.Equal(t, "custom-user", cfg.DBUser)
		assert.Equal(t, "custom-pass", cfg.DBPass)
		assert.Equal(t, "custom-db", cfg.DBName)
		assert.Equal(t, "require", cfg.DBSSLMode)
		assert.Equal(t, "custom-jwt-secret-key-that-is-long-enough-for-security-validation", cfg.JWTSecret)
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("InvalidServerPort", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "invalid",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "user",
			DBName:     "db",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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
			DBUser:     "user",
			DBName:     "db",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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
			DBUser:     "user",
			DBName:     "db",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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
			DBName:     "db",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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
			DBUser:     "user",
			DBName:     "",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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
			DBUser:     "user",
			DBName:     "db",
			JWTSecret:  "",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret is required")
	})

	t.Run("ShortJWTSecret", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "user",
			DBName:     "db",
			JWTSecret:  "short", // Too short
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret must be at least 32 characters")
	})

	t.Run("InvalidSSLMode", func(t *testing.T) {
		cfg := &Config{
			ServerPort: "8080",
			DBHost:     "localhost",
			DBPort:     "5432",
			DBUser:     "user",
			DBName:     "db",
			DBSSLMode:  "invalid",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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
			DBUser:     "user",
			DBName:     "db",
			DBSSLMode:  "disable",
			JWTSecret:  "test-jwt-secret-key-that-is-long-enough-for-validation",
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

	t.Run("EmptyEnvVar", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		result := getEnv("EMPTY_VAR", "default_value")
		assert.Equal(t, "", result)
	})
}
