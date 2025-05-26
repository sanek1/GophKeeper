package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

func LoadConfig() (*Config, error) {
	// Load .env file if exists, ignore error if not
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8081"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "admin"),
		DBPass:     getEnv("DB_PASS", "admin"),
		DBName:     getEnv("DB_NAME", "gophkeeper"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", DefaultJWTSecret),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// check server port
	if port, err := strconv.Atoi(c.ServerPort); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid server port: %s", c.ServerPort)
	}

	// check database port
	if port, err := strconv.Atoi(c.DBPort); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid database port: %s", c.DBPort)
	}

	// check required fields
	if c.DBHost == "" {
		return fmt.Errorf("database host is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("database user is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	// check SSL mode
	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[c.DBSSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", c.DBSSLMode)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
