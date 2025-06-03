package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecryptData(t *testing.T) {
	tests := []struct {
		name       string
		plaintext  []byte
		passphrase string
	}{
		{
			name:       "simple string",
			plaintext:  []byte("this is a secret message"),
			passphrase: "password123",
		},
		{
			name:       "empty data",
			plaintext:  []byte(""),
			passphrase: "password123",
		},
		{
			name:       "binary data",
			plaintext:  []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			passphrase: "secret_key",
		},
		{
			name:       "long text",
			plaintext:  bytes.Repeat([]byte("abvgdezik"), 100),
			passphrase: "very_long_password_for_testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// encrypt data
			encrypted, err := EncryptData(tt.plaintext, tt.passphrase)
			assert.NoError(t, err)
			assert.NotNil(t, encrypted)
			assert.NotEqual(t, tt.plaintext, encrypted)

			// decrypt data
			decrypted, err := DecryptData(encrypted, tt.passphrase)
			assert.NoError(t, err)

			// check length and content
			if len(tt.plaintext) == 0 {
				// for empty data check only length
				assert.Equal(t, 0, len(decrypted))
			} else {
				// for non-empty data check content
				assert.Equal(t, tt.plaintext, decrypted)
			}

			// check with wrong passphrase
			_, err = DecryptData(encrypted, tt.passphrase+"wrong")
			assert.Error(t, err)
		})
	}
}

func TestEncryptData_Error(t *testing.T) {
	// Test that encryption works even with empty password
	encrypted, err := EncryptData([]byte("test"), "")
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
}

func TestDecryptData_Error(t *testing.T) {
	// Test with too short encrypted data
	_, err := DecryptData([]byte("short"), "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")

	// Test with invalid encrypted data of sufficient length but corrupted
	encrypted, _ := EncryptData([]byte("test"), "password")
	if len(encrypted) > 0 {
		// Corrupt the last byte
		corrupted := make([]byte, len(encrypted))
		copy(corrupted, encrypted)
		corrupted[len(corrupted)-1] = corrupted[len(corrupted)-1] ^ 0xFF
		_, err = DecryptData(corrupted, "password")
		assert.Error(t, err)
	}
}

func TestEncryptData_EdgeCases(t *testing.T) {
	t.Run("EmptyData", func(t *testing.T) {
		data := []byte{}
		password := "testpassword"

		encrypted, err := EncryptData(data, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := DecryptData(encrypted, password)
		assert.NoError(t, err)
		assert.Empty(t, decrypted) // Empty data becomes nil after decryption
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		data := []byte("test data")
		password := ""

		encrypted, err := EncryptData(data, password)
		assert.NoError(t, err) // Empty password doesn't cause error, just weak encryption
		assert.NotEmpty(t, encrypted)
	})

	t.Run("VeryLongData", func(t *testing.T) {
		data := make([]byte, 1024*1024) // 1MB of data
		for i := range data {
			data[i] = byte(i % 256)
		}
		password := "testpassword"

		encrypted, err := EncryptData(data, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := DecryptData(encrypted, password)
		assert.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})
}

func TestDecryptData_EdgeCases(t *testing.T) {
	t.Run("InvalidData", func(t *testing.T) {
		invalidData := []byte("invalid encrypted data")
		password := "testpassword"

		decrypted, err := DecryptData(invalidData, password)
		assert.Error(t, err)
		assert.Nil(t, decrypted)
	})

	t.Run("TooShortData", func(t *testing.T) {
		shortData := []byte("short")
		password := "testpassword"

		decrypted, err := DecryptData(shortData, password)
		assert.Error(t, err)
		assert.Nil(t, decrypted)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		data := []byte("test data")
		correctPassword := "correctpassword"
		wrongPassword := "wrongpassword"

		encrypted, err := EncryptData(data, correctPassword)
		assert.NoError(t, err)

		decrypted, err := DecryptData(encrypted, wrongPassword)
		assert.Error(t, err)
		assert.Nil(t, decrypted)
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		data := []byte("test data")
		password := "testpassword"

		encrypted, err := EncryptData(data, password)
		assert.NoError(t, err)

		decrypted, err := DecryptData(encrypted, "")
		assert.Error(t, err)
		assert.Nil(t, decrypted)
	})
}

func TestGenerateKey_Coverage(t *testing.T) {
	t.Run("DifferentPasswords", func(t *testing.T) {
		key1 := generateKey("password1")
		assert.Len(t, key1, 32)

		key2 := generateKey("password2")
		assert.Len(t, key2, 32)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("SameSaltDifferentPasswords", func(t *testing.T) {
		key1 := generateKey("samepassword")
		key2 := generateKey("samepassword")

		assert.Equal(t, key1, key2) // Should be identical with same password
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		key := generateKey("")
		assert.Len(t, key, 32)
		assert.NotNil(t, key)
	})
}
