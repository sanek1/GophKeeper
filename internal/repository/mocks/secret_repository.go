package mocks

import (
	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/models"
	"github.com/yourusername/gophkeeper/internal/repository"
)

// MockSecretRepository реализует интерфейс репозитория секретов для тестов
type MockSecretRepository struct {
	secrets map[uuid.UUID]*models.Secret
}

// Убеждаемся, что MockSecretRepository реализует интерфейс SecretRepository
var _ repository.SecretRepository = (*MockSecretRepository)(nil)

// NewMockSecretRepository создает новый мок-репозиторий секретов
func NewMockSecretRepository() *MockSecretRepository {
	return &MockSecretRepository{
		secrets: make(map[uuid.UUID]*models.Secret),
	}
}

// Create создает новый секрет
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

// GetByID возвращает секрет по ID
func (r *MockSecretRepository) GetByID(id uuid.UUID) (*models.Secret, error) {
	if secret, ok := r.secrets[id]; ok {
		return secret, nil
	}
	return nil, nil
}

// GetByUserID возвращает все секреты пользователя
func (r *MockSecretRepository) GetByUserID(userID uuid.UUID) ([]*models.Secret, error) {
	var secrets []*models.Secret
	for _, secret := range r.secrets {
		if secret.UserID == userID {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

// Update обновляет существующий секрет
func (r *MockSecretRepository) Update(id uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	if secret, ok := r.secrets[id]; ok {
		secret.Type = secretType
		secret.Data = data
		secret.Metadata = metadata
		return secret, nil
	}
	return nil, nil
}

// Delete удаляет секрет
func (r *MockSecretRepository) Delete(id uuid.UUID) error {
	delete(r.secrets, id)
	return nil
}
