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
