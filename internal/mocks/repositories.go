package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository реализует интерфейс репозитория пользователей для тестов
type MockUserRepository struct {
	users map[string]*models.User
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

// Create creates a new user
func (r *MockUserRepository) Create(login, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        uuid.New(),
		Login:     login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	r.users[login] = user
	return user, nil
}

// GetByLogin returns a user by login
func (r *MockUserRepository) GetByLogin(login string) (*models.User, error) {
	if user, ok := r.users[login]; ok {
		return user, nil
	}
	return nil, nil
}

// GetByID returns a user by ID
func (r *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

// ValidatePassword checks the user's password
func (r *MockUserRepository) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// MockSecretRepository implements the secret repository interface for tests
type MockSecretRepository struct {
	secrets map[uuid.UUID]*models.Secret
}

// NewMockSecretRepository creates a new mock secret repository
func NewMockSecretRepository() *MockSecretRepository {
	return &MockSecretRepository{
		secrets: make(map[uuid.UUID]*models.Secret),
	}
}

// Create creates a new secret
func (r *MockSecretRepository) Create(userID uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	secret := &models.Secret{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      secretType,
		Data:      data,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

// GetByUserID returns all user's secrets
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
		secret.UpdatedAt = time.Now()
		return secret, nil
	}
	return nil, nil
}

// Delete deletes a secret
func (r *MockSecretRepository) Delete(id uuid.UUID) error {
	delete(r.secrets, id)
	return nil
}
