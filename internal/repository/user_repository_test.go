package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/config"
	"github.com/sanek1/GophKeeper/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	cfg := &config.Config{
		DBHost:    "localhost",
		DBPort:    "5432",
		DBUser:    "postgres",
		DBPass:    "admin",
		DBName:    "gophkeeper",
		DBSSLMode: "disable",
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		t.Skip("Database not available:", err)
	}

	// Create a users table for tests
	createUserTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err = db.DB().Exec(createUserTableQuery)
	require.NoError(t, err)

	// Clear the table before tests
	_, err = db.DB().Exec("DELETE FROM users")
	require.NoError(t, err)

	return db.DB()
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	t.Run("CreateUser", func(t *testing.T) {
		login := "test@example.com"
		password := "password123"

		user, err := repo.Create(login, password)
		require.NoError(t, err)
		require.NotNil(t, user)

		assert.Equal(t, login, user.Login)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.Password)
		assert.NotEqual(t, password, user.Password) // Password should be hashed
		assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
	})

	t.Run("CreateDuplicateLogin", func(t *testing.T) {
		login := "duplicate@example.com"
		password := "password123"

		// Create the first user
		_, err := repo.Create(login, password)
		require.NoError(t, err)

		// Try to create a second user with the same login
		_, err = repo.Create(login, password)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	t.Run("GetExistingUser", func(t *testing.T) {
		login := "gettest@example.com"
		password := "password123"

		// Create the user
		createdUser, err := repo.Create(login, password)
		require.NoError(t, err)

		// Get the user by login
		foundUser, err := repo.GetByLogin(login)
		require.NoError(t, err)
		require.NotNil(t, foundUser)

		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, createdUser.Login, foundUser.Login)
		assert.Equal(t, createdUser.Password, foundUser.Password)
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		foundUser, err := repo.GetByLogin("nonexistent@example.com")
		require.NoError(t, err)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	t.Run("GetExistingUser", func(t *testing.T) {
		login := "getbyid@example.com"
		password := "password123"

		// Create the user
		createdUser, err := repo.Create(login, password)
		require.NoError(t, err)

		// Get the user by ID
		foundUser, err := repo.GetByID(createdUser.ID)
		require.NoError(t, err)
		require.NotNil(t, foundUser)

		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, createdUser.Login, foundUser.Login)
		assert.Equal(t, createdUser.Password, foundUser.Password)
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		randomID := uuid.New()
		foundUser, err := repo.GetByID(randomID)
		require.NoError(t, err)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_ValidatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	login := "validate@example.com"
	password := "password123"

	// Create the user
	user, err := repo.Create(login, password)
	require.NoError(t, err)

	t.Run("ValidPassword", func(t *testing.T) {
		isValid := repo.ValidatePassword(user, password)
		assert.True(t, isValid)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		isValid := repo.ValidatePassword(user, "wrongpassword")
		assert.False(t, isValid)
	})
}
