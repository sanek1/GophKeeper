package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		user := User{
			ID:        uuid.New(),
			Login:     "test@example.com",
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "test@example.com", user.Login)
		assert.Equal(t, "hashedpassword", user.Password)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("UserJSONSerialization", func(t *testing.T) {
		user := User{
			ID:        uuid.New(),
			Login:     "json@example.com",
			Password:  "password",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		data, err := json.Marshal(user)
		require.NoError(t, err)

		var deserializedUser User
		err = json.Unmarshal(data, &deserializedUser)
		require.NoError(t, err)

		assert.Equal(t, user.ID, deserializedUser.ID)
		assert.Equal(t, user.Login, deserializedUser.Login)
		assert.Empty(t, deserializedUser.Password)
	})
}

func TestSecret(t *testing.T) {
	t.Run("CreateSecret", func(t *testing.T) {
		userID := uuid.New()
		secret := Secret{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      "password",
			Data:      []byte("encrypted_data"),
			Metadata:  "Test Password",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NotEmpty(t, secret.ID)
		assert.Equal(t, userID, secret.UserID)
		assert.Equal(t, "password", secret.Type)
		assert.Equal(t, []byte("encrypted_data"), secret.Data)
		assert.Equal(t, "Test Password", secret.Metadata)
		assert.False(t, secret.CreatedAt.IsZero())
		assert.False(t, secret.UpdatedAt.IsZero())
	})

	t.Run("SecretJSONSerialization", func(t *testing.T) {
		secret := Secret{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Type:      "note",
			Data:      []byte("secret note content"),
			Metadata:  "Important Note",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		data, err := json.Marshal(secret)
		require.NoError(t, err)

		var deserializedSecret Secret
		err = json.Unmarshal(data, &deserializedSecret)
		require.NoError(t, err)

		assert.Equal(t, secret.ID, deserializedSecret.ID)
		assert.Equal(t, secret.UserID, deserializedSecret.UserID)
		assert.Equal(t, secret.Type, deserializedSecret.Type)
		assert.Equal(t, secret.Data, deserializedSecret.Data)
		assert.Equal(t, secret.Metadata, deserializedSecret.Metadata)
	})
}

func TestSecretTypes(t *testing.T) {
	t.Run("ValidSecretTypes", func(t *testing.T) {
		expectedTypes := []string{"password", "card", "text", "file", "note", "binary"}
		assert.Equal(t, expectedTypes, SecretTypes)
		assert.Len(t, SecretTypes, 6)
	})

	t.Run("CheckSecretTypeExists", func(t *testing.T) {
		validTypes := map[string]bool{
			"password": true,
			"card":     true,
			"text":     true,
			"file":     true,
			"note":     true,
			"binary":   true,
		}

		for _, secretType := range SecretTypes {
			assert.True(t, validTypes[secretType], "Secret type %s should be valid", secretType)
		}
	})
}

func TestPasswordData(t *testing.T) {
	t.Run("CreatePasswordData", func(t *testing.T) {
		passwordData := PasswordData{
			Website:  "https://example.com",
			Username: "testuser",
			Password: "testpassword",
			Notes:    "Test notes",
		}

		assert.Equal(t, "https://example.com", passwordData.Website)
		assert.Equal(t, "testuser", passwordData.Username)
		assert.Equal(t, "testpassword", passwordData.Password)
		assert.Equal(t, "Test notes", passwordData.Notes)
	})

	t.Run("PasswordDataJSONSerialization", func(t *testing.T) {
		passwordData := PasswordData{
			Website:  "https://test.com",
			Username: "user123",
			Password: "pass123",
			Notes:    "Important notes",
		}

		data, err := json.Marshal(passwordData)
		require.NoError(t, err)

		var deserializedData PasswordData
		err = json.Unmarshal(data, &deserializedData)
		require.NoError(t, err)

		assert.Equal(t, passwordData.Website, deserializedData.Website)
		assert.Equal(t, passwordData.Username, deserializedData.Username)
		assert.Equal(t, passwordData.Password, deserializedData.Password)
		assert.Equal(t, passwordData.Notes, deserializedData.Notes)
	})
}

func TestCardData(t *testing.T) {
	t.Run("CreateCardData", func(t *testing.T) {
		cardData := CardData{
			CardNumber:  "1234567890123456",
			ExpiryMonth: "12",
			ExpiryYear:  "25",
			CVV:         "123",
			CardHolder:  "John Doe",
			BankName:    "Test Bank",
			CardType:    "Visa",
		}

		assert.Equal(t, "1234567890123456", cardData.CardNumber)
		assert.Equal(t, "12", cardData.ExpiryMonth)
		assert.Equal(t, "25", cardData.ExpiryYear)
		assert.Equal(t, "123", cardData.CVV)
		assert.Equal(t, "John Doe", cardData.CardHolder)
		assert.Equal(t, "Test Bank", cardData.BankName)
		assert.Equal(t, "Visa", cardData.CardType)
	})

	t.Run("CardDataJSONSerialization", func(t *testing.T) {
		cardData := CardData{
			CardNumber:  "9876543210987654",
			ExpiryMonth: "06",
			ExpiryYear:  "27",
			CVV:         "456",
			CardHolder:  "Jane Smith",
			BankName:    "Another Bank",
			CardType:    "MasterCard",
		}

		data, err := json.Marshal(cardData)
		require.NoError(t, err)

		var deserializedData CardData
		err = json.Unmarshal(data, &deserializedData)
		require.NoError(t, err)

		assert.Equal(t, cardData.CardNumber, deserializedData.CardNumber)
		assert.Equal(t, cardData.ExpiryMonth, deserializedData.ExpiryMonth)
		assert.Equal(t, cardData.ExpiryYear, deserializedData.ExpiryYear)
		assert.Equal(t, cardData.CVV, deserializedData.CVV)
		assert.Equal(t, cardData.CardHolder, deserializedData.CardHolder)
		assert.Equal(t, cardData.BankName, deserializedData.BankName)
		assert.Equal(t, cardData.CardType, deserializedData.CardType)
	})
}

func TestRegisterRequest(t *testing.T) {
	t.Run("CreateRegisterRequest", func(t *testing.T) {
		req := RegisterRequest{
			Login:    "newuser@example.com",
			Password: "newpassword",
		}

		assert.Equal(t, "newuser@example.com", req.Login)
		assert.Equal(t, "newpassword", req.Password)
	})

	t.Run("RegisterRequestJSONSerialization", func(t *testing.T) {
		req := RegisterRequest{
			Login:    "register@test.com",
			Password: "registerpass",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var deserializedReq RegisterRequest
		err = json.Unmarshal(data, &deserializedReq)
		require.NoError(t, err)

		assert.Equal(t, req.Login, deserializedReq.Login)
		assert.Equal(t, req.Password, deserializedReq.Password)
	})
}

func TestLoginRequest(t *testing.T) {
	t.Run("CreateLoginRequest", func(t *testing.T) {
		req := LoginRequest{
			Login:    "user@example.com",
			Password: "userpassword",
		}

		assert.Equal(t, "user@example.com", req.Login)
		assert.Equal(t, "userpassword", req.Password)
	})

	t.Run("LoginRequestJSONSerialization", func(t *testing.T) {
		req := LoginRequest{
			Login:    "login@test.com",
			Password: "loginpass",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var deserializedReq LoginRequest
		err = json.Unmarshal(data, &deserializedReq)
		require.NoError(t, err)

		assert.Equal(t, req.Login, deserializedReq.Login)
		assert.Equal(t, req.Password, deserializedReq.Password)
	})
}

func TestSecretRequest(t *testing.T) {
	t.Run("CreateSecretRequest", func(t *testing.T) {
		req := SecretRequest{
			Type:     "password",
			Data:     []byte("encrypted_secret_data"),
			Metadata: "Test Secret",
		}

		assert.Equal(t, "password", req.Type)
		assert.Equal(t, []byte("encrypted_secret_data"), req.Data)
		assert.Equal(t, "Test Secret", req.Metadata)
	})

	t.Run("SecretRequestJSONSerialization", func(t *testing.T) {
		req := SecretRequest{
			Type:     "note",
			Data:     []byte("note content"),
			Metadata: "Important Note",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var deserializedReq SecretRequest
		err = json.Unmarshal(data, &deserializedReq)
		require.NoError(t, err)

		assert.Equal(t, req.Type, deserializedReq.Type)
		assert.Equal(t, req.Data, deserializedReq.Data)
		assert.Equal(t, req.Metadata, deserializedReq.Metadata)
	})
}

func TestClientInfo(t *testing.T) {
	t.Run("CreateClientInfo", func(t *testing.T) {
		info := ClientInfo{
			Version:     "1.0.0",
			BuildDate:   "2025-01-01",
			CommitHash:  "abc123",
			BuildNumber: "42",
		}

		assert.Equal(t, "1.0.0", info.Version)
		assert.Equal(t, "2025-01-01", info.BuildDate)
		assert.Equal(t, "abc123", info.CommitHash)
		assert.Equal(t, "42", info.BuildNumber)
	})

	t.Run("ClientInfoJSONSerialization", func(t *testing.T) {
		info := ClientInfo{
			Version:     "2.1.0",
			BuildDate:   "2025-02-01",
			CommitHash:  "def456",
			BuildNumber: "100",
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		var deserializedInfo ClientInfo
		err = json.Unmarshal(data, &deserializedInfo)
		require.NoError(t, err)

		assert.Equal(t, info.Version, deserializedInfo.Version)
		assert.Equal(t, info.BuildDate, deserializedInfo.BuildDate)
		assert.Equal(t, info.CommitHash, deserializedInfo.CommitHash)
		assert.Equal(t, info.BuildNumber, deserializedInfo.BuildNumber)
	})
}
