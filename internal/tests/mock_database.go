package tests

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/database"
	"github.com/yourusername/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// MockDatabase реализует интерфейс database.DBInterface для тестирования
type MockDatabase struct {
	users   map[string]*models.User
	secrets map[uuid.UUID]*models.Secret
}

// Убеждаемся, что MockDatabase реализует интерфейс DBInterface
var _ database.DBInterface = (*MockDatabase)(nil)

// NewMockDatabase создает новый экземпляр мок-базы данных
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		users:   make(map[string]*models.User),
		secrets: make(map[uuid.UUID]*models.Secret),
	}
}

// DB возвращает SQL-соединение
func (m *MockDatabase) DB() *sql.DB {
	return nil
}

// Close закрывает соединение с базой данных
func (m *MockDatabase) Close() error {
	return nil
}

// CreateUser создает нового пользователя
func (m *MockDatabase) CreateUser(login, password string) (*models.User, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем нового пользователя
	user := &models.User{
		ID:       uuid.New(),
		Login:    login,
		Password: string(hashedPassword),
	}

	// Сохраняем пользователя в мок-хранилище
	m.users[login] = user

	return user, nil
}

// GetUserByLogin возвращает пользователя по логину
func (m *MockDatabase) GetUserByLogin(login string) (*models.User, error) {
	if user, ok := m.users[login]; ok {
		return user, nil
	}
	return nil, nil
}

// GetUserByID возвращает пользователя по ID
func (m *MockDatabase) GetUserByID(id uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

// CreateSecret создает новый секрет
func (m *MockDatabase) CreateSecret(secret *models.Secret) error {
	secret.ID = uuid.New()
	m.secrets[secret.ID] = secret
	return nil
}

// GetSecrets возвращает все секреты пользователя
func (m *MockDatabase) GetSecrets(userID uuid.UUID) ([]*models.Secret, error) {
	var secrets []*models.Secret
	for _, secret := range m.secrets {
		if secret.UserID == userID {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

// GetSecret возвращает секрет по ID
func (m *MockDatabase) GetSecret(id uuid.UUID) (*models.Secret, error) {
	if secret, ok := m.secrets[id]; ok {
		return secret, nil
	}
	return nil, nil
}

// UpdateSecret обновляет существующий секрет
func (m *MockDatabase) UpdateSecret(secret *models.Secret) error {
	if _, ok := m.secrets[secret.ID]; ok {
		m.secrets[secret.ID] = secret
		return nil
	}
	return nil
}

// DeleteSecret удаляет секрет
func (m *MockDatabase) DeleteSecret(id uuid.UUID) error {
	delete(m.secrets, id)
	return nil
}
