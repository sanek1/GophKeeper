package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/models"
	"github.com/yourusername/gophkeeper/internal/repository"
)

// SecretRepositoryMock represents a mock secret repository for testing
type SecretRepositoryMock struct {
	Secrets map[uuid.UUID]*models.Secret // Secrets indexed by ID
}

// NewSecretRepositoryMock creates a new mock instance
func NewSecretRepositoryMock() *SecretRepositoryMock {
	return &SecretRepositoryMock{
		Secrets: make(map[uuid.UUID]*models.Secret),
	}
}

// Ensure SecretRepositoryMock implements SecretRepository
var _ repository.SecretRepository = (*SecretRepositoryMock)(nil)

// Create creates a new secret
func (m *SecretRepositoryMock) Create(userID uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	secret := &models.Secret{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      secretType,
		Data:      data,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.Secrets[secret.ID] = secret
	return secret, nil
}

// GetByID returns secret by ID
func (m *SecretRepositoryMock) GetByID(id uuid.UUID) (*models.Secret, error) {
	secret, exists := m.Secrets[id]
	if !exists {
		return nil, nil
	}
	return secret, nil
}

// GetByUserID returns all user secrets
func (m *SecretRepositoryMock) GetByUserID(userID uuid.UUID) ([]*models.Secret, error) {
	var secrets []*models.Secret
	for _, secret := range m.Secrets {
		if secret.UserID == userID {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

// Update updates existing secret
func (m *SecretRepositoryMock) Update(id uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	secret, exists := m.Secrets[id]
	if !exists {
		return nil, ErrSecretNotFound
	}

	secret.Type = secretType
	secret.Data = data
	secret.Metadata = metadata
	secret.UpdatedAt = time.Now()

	return secret, nil
}

// Delete deletes secret by ID
func (m *SecretRepositoryMock) Delete(id uuid.UUID) error {
	if _, exists := m.Secrets[id]; !exists {
		return ErrSecretNotFound
	}

	delete(m.Secrets, id)
	return nil
}

// AddTestSecret adds a test secret to the mock repository
func (m *SecretRepositoryMock) AddTestSecret(userID uuid.UUID, secretType string, data []byte, metadata string) *models.Secret {
	secret := &models.Secret{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      secretType,
		Data:      data,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.Secrets[secret.ID] = secret
	return secret
}
