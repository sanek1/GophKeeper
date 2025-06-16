package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSecretRepository(t *testing.T) {
	// Создаем мок репозитория
	repo := mocks.NewSecretRepositoryMock()

	// Тестовые данные
	userID := uuid.New()
	secretType := "password"
	secretData := []byte("encrypted_secret_data")
	metadata := "Пароль от почты"

	// Тест: Создание секрета
	t.Run("Create", func(t *testing.T) {
		secret, err := repo.Create(userID, secretType, secretData, metadata)
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, userID, secret.UserID)
		assert.Equal(t, secretType, secret.Type)
		assert.Equal(t, secretData, secret.Data)
		assert.Equal(t, metadata, secret.Metadata)
	})

	// Тест: Получение секрета по ID
	t.Run("GetByID", func(t *testing.T) {
		// Создаем секрет для последующего получения
		createdSecret, _ := repo.Create(userID, secretType, secretData, metadata)

		// Получаем секрет по ID
		secret, err := repo.GetByID(createdSecret.ID)
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, createdSecret.ID, secret.ID)
		assert.Equal(t, userID, secret.UserID)
		assert.Equal(t, secretType, secret.Type)
		assert.Equal(t, secretData, secret.Data)
		assert.Equal(t, metadata, secret.Metadata)
	})

	// Тест: Получение несуществующего секрета
	t.Run("GetByID_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New()
		secret, err := repo.GetByID(nonExistentID)
		assert.NoError(t, err)
		assert.Nil(t, secret)
	})

	// Тест: Получение секретов пользователя
	t.Run("GetByUserID", func(t *testing.T) {
		// Создаем новый мок репозиторий для этого теста, чтобы не зависеть от предыдущих тестов
		newRepo := mocks.NewSecretRepositoryMock()

		// Создаем секреты для пользователя
		newRepo.Create(userID, "password", []byte("data1"), "Пароль 1")
		newRepo.Create(userID, "card", []byte("data2"), "Карта 1")
		newRepo.Create(userID, "note", []byte("data3"), "Заметка 1")

		// Создаем секрет для другого пользователя
		anotherUserID := uuid.New()
		newRepo.Create(anotherUserID, "password", []byte("other_data"), "Другой пароль")

		// Получаем секреты пользователя
		secrets, err := newRepo.GetByUserID(userID)
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
		assert.Equal(t, 3, len(secrets)) // Должно быть ровно 3 секрета для нашего пользователя

		// Проверяем, что все секреты принадлежат пользователю
		for _, secret := range secrets {
			assert.Equal(t, userID, secret.UserID)
		}
	})

	// Тест: Обновление секрета
	t.Run("Update", func(t *testing.T) {
		// Создаем секрет для последующего обновления
		createdSecret, _ := repo.Create(userID, secretType, secretData, metadata)

		// Новые данные для обновления
		newType := "card"
		newData := []byte("new_encrypted_data")
		newMetadata := "Обновленная карта"

		// Обновляем секрет
		updatedSecret, err := repo.Update(createdSecret.ID, newType, newData, newMetadata)
		assert.NoError(t, err)
		assert.NotNil(t, updatedSecret)
		assert.Equal(t, createdSecret.ID, updatedSecret.ID)
		assert.Equal(t, userID, updatedSecret.UserID)
		assert.Equal(t, newType, updatedSecret.Type)
		assert.Equal(t, newData, updatedSecret.Data)
		assert.Equal(t, newMetadata, updatedSecret.Metadata)
	})

	// Тест: Обновление несуществующего секрета
	t.Run("Update_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.Update(nonExistentID, "password", []byte("data"), "metadata")
		assert.Error(t, err)
		assert.Equal(t, mocks.ErrSecretNotFound, err)
	})

	// Тест: Удаление секрета
	t.Run("Delete", func(t *testing.T) {
		// Создаем секрет для последующего удаления
		createdSecret, _ := repo.Create(userID, secretType, secretData, metadata)

		// Удаляем секрет
		err := repo.Delete(createdSecret.ID)
		assert.NoError(t, err)

		// Проверяем, что секрет удален
		secret, _ := repo.GetByID(createdSecret.ID)
		assert.Nil(t, secret)
	})

	// Тест: Удаление несуществующего секрета
	t.Run("Delete_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := repo.Delete(nonExistentID)
		assert.Error(t, err)
		assert.Equal(t, mocks.ErrSecretNotFound, err)
	})
}
