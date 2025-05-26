package mocks

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/database"
	"github.com/yourusername/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// MockDatabase implements database.DBInterface for testing
type MockDatabase struct {
	users   map[string]*models.User
	secrets map[uuid.UUID]*models.Secret
}

// Ensure that MockDatabase implements DBInterface
var _ database.DBInterface = (*MockDatabase)(nil)

// NewMockDatabase creates a new instance of mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		users:   make(map[string]*models.User),
		secrets: make(map[uuid.UUID]*models.Secret),
	}
}

// DB returns SQL connection
func (m *MockDatabase) DB() *sql.DB {
	return nil
}

// Close closes database connection
func (m *MockDatabase) Close() error {
	return nil
}

// CreateUser creates a new user
func (m *MockDatabase) CreateUser(login, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       uuid.New(),
		Login:    login,
		Password: string(hashedPassword),
	}

	m.users[login] = user
	return user, nil
}

// GetUserByLogin returns user by login
func (m *MockDatabase) GetUserByLogin(login string) (*models.User, error) {
	if user, ok := m.users[login]; ok {
		return user, nil
	}
	return nil, nil
}

// GetUserByID returns user by ID
func (m *MockDatabase) GetUserByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

// CreateSecret creates a new secret
func (m *MockDatabase) CreateSecret(secret *models.Secret) error {
	secret.ID = uuid.New()
	m.secrets[secret.ID] = secret
	return nil
}

// GetSecrets returns all user secrets
func (m *MockDatabase) GetSecrets(userID uuid.UUID) ([]*models.Secret, error) {
	var secrets []*models.Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

// GetSecret returns secret by ID
func (m *MockDatabase) GetSecret(id uuid.UUID) (*models.Secret, error) {
	if secret, ok := m.secrets[id]; ok {
		return secret, nil
	}
	return nil, nil
}

// UpdateSecret updates existing secret
func (m *MockDatabase) UpdateSecret(secret *models.Secret) error {
	if _, ok := m.secrets[secret.ID]; ok {
		m.secrets[secret.ID] = secret
		return nil
	}
	return nil
}

// DeleteSecret deletes secret
func (m *MockDatabase) DeleteSecret(id uuid.UUID) error {
	delete(m.secrets, id)
	return nil
}
