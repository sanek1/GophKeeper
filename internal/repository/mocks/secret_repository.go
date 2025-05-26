package mocks

import (
	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/sanek1/GophKeeper/internal/repository"
)

// MockSecretRepository implements the SecretRepository interface for testing
type MockSecretRepository struct {
	secrets map[uuid.UUID]*models.Secret
}

// ensure that MockSecretRepository implements the SecretRepository interface
var _ repository.SecretRepository = (*MockSecretRepository)(nil)

// NewMockSecretRepository creates a new mock secret repository
func NewMockSecretRepository() *MockSecretRepository {
	return &MockSecretRepository{
		secrets: make(map[uuid.UUID]*models.Secret),
	}
}

// Create creates a new secret
func (r *MockSecretRepository) Create(userID uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	secret := &models.Secret{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     secretType,
		Data:     data,
		Metadata: metadata,
	}

	r.secrets[secret.ID] = secret
	return secret, nil
}

// GetByID returns a secret by ID
func (r *MockSecretRepository) GetByID(id uuid.UUID) (*models.Secret, error) {
	if secret, ok := r.secrets[id]; ok {
		return secret, nil
	}
	return nil, nil
}

// GetByUserID returns all secrets for a user
func (r *MockSecretRepository) GetByUserID(userID uuid.UUID) ([]*models.Secret, error) {
	var secrets []*models.Secret
	for _, secret := range r.secrets {
		if secret.UserID == userID {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

// Update updates an existing secret
func (r *MockSecretRepository) Update(id uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	if secret, ok := r.secrets[id]; ok {
		secret.Type = secretType
		secret.Data = data
		secret.Metadata = metadata
		return secret, nil
	}
	return nil, nil
}

// Delete deletes a secret
func (r *MockSecretRepository) Delete(id uuid.UUID) error {
	delete(r.secrets, id)
	return nil
}
