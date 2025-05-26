package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSecretDB(t *testing.T) (*sql.DB, uuid.UUID) {
	db := setupTestDB(t)

	// Create secrets table for tests
	createSecretTableQuery := `
		CREATE TABLE IF NOT EXISTS secrets (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			type VARCHAR(50) NOT NULL,
			data BYTEA NOT NULL,
			metadata TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err := db.Exec(createSecretTableQuery)
	require.NoError(t, err)

	// Clear the table before tests
	_, err = db.Exec("DELETE FROM secrets")
	require.NoError(t, err)

	// Create a test user
	userRepo := NewUserRepository(db)
	user, err := userRepo.Create("secrettest@example.com", "password123")
	require.NoError(t, err)

	return db, user.ID
}

func TestSecretRepository_Create(t *testing.T) {
	db, userID := setupTestSecretDB(t)
	repo := NewSecretRepository(db)

	t.Run("CreateSecret", func(t *testing.T) {
		secretType := "password"
		data := []byte("encrypted_password_data")
		metadata := "Test Password"

		secret, err := repo.Create(userID, secretType, data, metadata)
		require.NoError(t, err)
		require.NotNil(t, secret)

		assert.Equal(t, userID, secret.UserID)
		assert.Equal(t, secretType, secret.Type)
		assert.Equal(t, data, secret.Data)
		assert.Equal(t, metadata, secret.Metadata)
		assert.NotEmpty(t, secret.ID)
		assert.WithinDuration(t, time.Now(), secret.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), secret.UpdatedAt, time.Second)
	})
}

func TestSecretRepository_GetByID(t *testing.T) {
	db, userID := setupTestSecretDB(t)
	repo := NewSecretRepository(db)

	t.Run("GetExistingSecret", func(t *testing.T) {
		// Create a secret
		secretType := "note"
		data := []byte("secret note content")
		metadata := "Important Note"

		createdSecret, err := repo.Create(userID, secretType, data, metadata)
		require.NoError(t, err)

		// Get the secret by ID
		foundSecret, err := repo.GetByID(createdSecret.ID)
		require.NoError(t, err)
		require.NotNil(t, foundSecret)

		assert.Equal(t, createdSecret.ID, foundSecret.ID)
		assert.Equal(t, createdSecret.UserID, foundSecret.UserID)
		assert.Equal(t, createdSecret.Type, foundSecret.Type)
		assert.Equal(t, createdSecret.Data, foundSecret.Data)
		assert.Equal(t, createdSecret.Metadata, foundSecret.Metadata)
	})

	t.Run("GetNonExistentSecret", func(t *testing.T) {
		randomID := uuid.New()
		foundSecret, err := repo.GetByID(randomID)
		require.NoError(t, err)
		assert.Nil(t, foundSecret)
	})
}

func TestSecretRepository_GetByUserID(t *testing.T) {
	db, userID := setupTestSecretDB(t)
	repo := NewSecretRepository(db)

	t.Run("GetSecretsForUser", func(t *testing.T) {
		// Create several secrets for the user
		secrets := []struct {
			Type     string
			Data     []byte
			Metadata string
		}{
			{"password", []byte("password1"), "Site 1"},
			{"card", []byte("card_data"), "Bank Card"},
			{"note", []byte("note_content"), "Personal Note"},
		}

		createdSecrets := make([]*models.Secret, 0, len(secrets))
		for _, s := range secrets {
			secret, err := repo.Create(userID, s.Type, s.Data, s.Metadata)
			require.NoError(t, err)
			createdSecrets = append(createdSecrets, secret)
		}

		// Get all secrets for the user
		foundSecrets, err := repo.GetByUserID(userID)
		require.NoError(t, err)
		assert.Len(t, foundSecrets, len(secrets))

		// Check that all created secrets are found
		for _, created := range createdSecrets {
			found := false
			for _, found_secret := range foundSecrets {
				if created.ID == found_secret.ID {
					found = true
					assert.Equal(t, created.Type, found_secret.Type)
					assert.Equal(t, created.Data, found_secret.Data)
					assert.Equal(t, created.Metadata, found_secret.Metadata)
					break
				}
			}
			assert.True(t, found, "Created secret not found in results")
		}
	})

	t.Run("GetSecretsForNonExistentUser", func(t *testing.T) {
		randomUserID := uuid.New()
		foundSecrets, err := repo.GetByUserID(randomUserID)
		require.NoError(t, err)
		assert.Empty(t, foundSecrets)
	})
}

func TestSecretRepository_Update(t *testing.T) {
	db, userID := setupTestSecretDB(t)
	repo := NewSecretRepository(db)

	t.Run("UpdateExistingSecret", func(t *testing.T) {
		// Create a secret
		originalType := "password"
		originalData := []byte("original_data")
		originalMetadata := "Original"

		createdSecret, err := repo.Create(userID, originalType, originalData, originalMetadata)
		require.NoError(t, err)

		// Update the secret
		newType := "text"
		newData := []byte("updated_data")
		newMetadata := "Updated"

		updatedSecret, err := repo.Update(createdSecret.ID, newType, newData, newMetadata)
		require.NoError(t, err)
		require.NotNil(t, updatedSecret)

		assert.Equal(t, createdSecret.ID, updatedSecret.ID)
		assert.Equal(t, createdSecret.UserID, updatedSecret.UserID)
		assert.Equal(t, newType, updatedSecret.Type)
		assert.Equal(t, newData, updatedSecret.Data)
		assert.Equal(t, newMetadata, updatedSecret.Metadata)
		assert.True(t, updatedSecret.UpdatedAt.After(createdSecret.UpdatedAt))
	})

	t.Run("UpdateNonExistentSecret", func(t *testing.T) {
		randomID := uuid.New()
		_, err := repo.Update(randomID, "text", []byte("data"), "metadata")
		assert.Error(t, err)
	})
}

func TestSecretRepository_Delete(t *testing.T) {
	db, userID := setupTestSecretDB(t)
	repo := NewSecretRepository(db)

	t.Run("DeleteExistingSecret", func(t *testing.T) {
		// Create a secret
		secret, err := repo.Create(userID, "password", []byte("data"), "metadata")
		require.NoError(t, err)

		// Delete the secret
		err = repo.Delete(secret.ID)
		require.NoError(t, err)

		// Check that the secret is deleted
		foundSecret, err := repo.GetByID(secret.ID)
		require.NoError(t, err)
		assert.Nil(t, foundSecret)
	})

	t.Run("DeleteNonExistentSecret", func(t *testing.T) {
		randomID := uuid.New()
		err := repo.Delete(randomID)
		// Deleting a non-existent secret should not cause an error
		require.NoError(t, err)
	})
}
