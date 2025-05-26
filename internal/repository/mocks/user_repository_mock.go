package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/models"
	"github.com/yourusername/gophkeeper/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserRepositoryMock represents a mock user repository for testing
type UserRepositoryMock struct {
	Users map[string]*models.User // Users indexed by login
}

// NewUserRepositoryMock creates a new mock instance
func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{
		Users: make(map[string]*models.User),
	}
}

// Ensure UserRepositoryMock implements UserRepository
var _ repository.UserRepository = (*UserRepositoryMock)(nil)

// Create creates a new user
func (m *UserRepositoryMock) Create(login, password string) (*models.User, error) {
	// Check if user with this login already exists
	if _, exists := m.Users[login]; exists {
		return nil, ErrUserAlreadyExists
	}

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

	m.Users[login] = user
	return user, nil
}

// GetByLogin returns user by login
func (m *UserRepositoryMock) GetByLogin(login string) (*models.User, error) {
	user, exists := m.Users[login]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// GetByID returns user by ID
func (m *UserRepositoryMock) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

// ValidatePassword checks user password
func (m *UserRepositoryMock) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// AddTestUser adds a test user to the mock repository
func (m *UserRepositoryMock) AddTestUser(login, password string) (*models.User, error) {
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

	m.Users[login] = user
	return user, nil
}
