package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/sanek1/GophKeeper/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository represents a mock user repository for testing
type MockUserRepository struct {
	Users map[string]*models.User // Users indexed by login
}

// NewMockUserRepository creates a new mock instance
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*models.User),
	}
}

// Ensure MockUserRepository implements UserRepository
var _ repository.UserRepository = (*MockUserRepository)(nil)

// Create creates a new user
func (m *MockUserRepository) Create(login, password string) (*models.User, error) {
	// check if a user with this login already exists
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

// GetByLogin returns a user by login
func (m *MockUserRepository) GetByLogin(login string) (*models.User, error) {
	user, exists := m.Users[login]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// GetByID returns a user by ID
func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

// ValidatePassword checks the user's password
func (m *MockUserRepository) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
