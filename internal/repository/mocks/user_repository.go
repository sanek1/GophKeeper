package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/models"
	"github.com/yourusername/gophkeeper/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository представляет мок репозитория пользователей для тестирования
type MockUserRepository struct {
	Users map[string]*models.User // Пользователи, индексированные по логину
}

// NewMockUserRepository создает новый экземпляр мока
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*models.User),
	}
}

// Ensure MockUserRepository implements UserRepository
var _ repository.UserRepository = (*MockUserRepository)(nil)

// Create создает нового пользователя
func (m *MockUserRepository) Create(login, password string) (*models.User, error) {
	// Проверяем, что пользователь с таким логином еще не существует
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

// GetByLogin возвращает пользователя по логину
func (m *MockUserRepository) GetByLogin(login string) (*models.User, error) {
	user, exists := m.Users[login]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// GetByID возвращает пользователя по ID
func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

// ValidatePassword проверяет пароль пользователя
func (m *MockUserRepository) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
