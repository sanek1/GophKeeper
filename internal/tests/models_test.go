package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestModels(t *testing.T) {
	// Тестируем создание и валидацию данных модели User
	t.Run("User", func(t *testing.T) {
		user := &models.User{
			ID:        uuid.New(),
			Login:     "test@example.com",
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NotNil(t, user)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "test@example.com", user.Login)
		assert.Equal(t, "hashedpassword", user.Password)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	// Тестируем создание и валидацию данных модели Secret
	t.Run("Secret", func(t *testing.T) {
		secret := &models.Secret{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Type:      "password",
			Data:      []byte("encrypted data"),
			Metadata:  "Пароль от сервиса",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NotNil(t, secret)
		assert.NotEqual(t, uuid.Nil, secret.ID)
		assert.NotEqual(t, uuid.Nil, secret.UserID)
		assert.Equal(t, "password", secret.Type)
		assert.Equal(t, []byte("encrypted data"), secret.Data)
		assert.Equal(t, "Пароль от сервиса", secret.Metadata)
		assert.False(t, secret.CreatedAt.IsZero())
		assert.False(t, secret.UpdatedAt.IsZero())
	})

	// Тестируем валидацию запросов и структур данных
	t.Run("PasswordData", func(t *testing.T) {
		passwordData := &models.PasswordData{
			Username: "user123",
			Password: "secretpass",
			Website:  "example.com",
			Notes:    "Важные заметки",
		}

		assert.NotNil(t, passwordData)
		assert.Equal(t, "user123", passwordData.Username)
		assert.Equal(t, "secretpass", passwordData.Password)
		assert.Equal(t, "example.com", passwordData.Website)
		assert.Equal(t, "Важные заметки", passwordData.Notes)
	})

	t.Run("CardData", func(t *testing.T) {
		cardData := &models.CardData{
			CardNumber:  "1234 5678 9012 3456",
			CardHolder:  "IVAN IVANOV",
			ExpiryMonth: "12",
			ExpiryYear:  "2025",
			CVV:         "123",
			BankName:    "Sberbank",
			CardType:    "Visa",
			Notes:       "Основная карта",
		}

		assert.NotNil(t, cardData)
		assert.Equal(t, "1234 5678 9012 3456", cardData.CardNumber)
		assert.Equal(t, "IVAN IVANOV", cardData.CardHolder)
		assert.Equal(t, "12", cardData.ExpiryMonth)
		assert.Equal(t, "2025", cardData.ExpiryYear)
		assert.Equal(t, "123", cardData.CVV)
		assert.Equal(t, "Sberbank", cardData.BankName)
		assert.Equal(t, "Visa", cardData.CardType)
		assert.Equal(t, "Основная карта", cardData.Notes)
	})

	t.Run("SecretTypes", func(t *testing.T) {
		// Проверяем, что определены типы секретов
		assert.NotEmpty(t, models.SecretTypes)
		assert.Contains(t, models.SecretTypes, "password")
		assert.Contains(t, models.SecretTypes, "card")
		assert.Contains(t, models.SecretTypes, "text")
		assert.Contains(t, models.SecretTypes, "file")
		assert.Contains(t, models.SecretTypes, "note")
		assert.Contains(t, models.SecretTypes, "binary")
	})
}
