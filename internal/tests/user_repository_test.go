package tests

import (
	"testing"

	"github.com/sanek1/GophKeeper/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
	// Создаем мок репозитория
	repo := mocks.NewUserRepositoryMock()

	// Тестовые данные
	login := "test@example.com"
	password := "password123"

	// Тест: Создание пользователя
	t.Run("Create", func(t *testing.T) {
		user, err := repo.Create(login, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, login, user.Login)
		assert.NotEmpty(t, user.Password)
		assert.NotEqual(t, password, user.Password) // Пароль должен быть хеширован
	})

	// Тест: Попытка создать пользователя с существующим логином
	t.Run("Create_DuplicateLogin", func(t *testing.T) {
		_, err := repo.Create(login, "anotherpassword")
		assert.Error(t, err)
		assert.Equal(t, mocks.ErrUserAlreadyExists, err)
	})

	// Тест: Получение пользователя по логину
	t.Run("GetByLogin", func(t *testing.T) {
		user, err := repo.GetByLogin(login)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, login, user.Login)
	})

	// Тест: Получение несуществующего пользователя
	t.Run("GetByLogin_NotFound", func(t *testing.T) {
		user, err := repo.GetByLogin("nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	// Тест: Проверка правильного пароля
	t.Run("ValidatePassword_Correct", func(t *testing.T) {
		user, _ := repo.GetByLogin(login)
		isValid := repo.ValidatePassword(user, password)
		assert.True(t, isValid)
	})

	// Тест: Проверка неправильного пароля
	t.Run("ValidatePassword_Incorrect", func(t *testing.T) {
		user, _ := repo.GetByLogin(login)
		isValid := repo.ValidatePassword(user, "wrongpassword")
		assert.False(t, isValid)
	})
}
