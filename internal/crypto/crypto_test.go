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
	// errors during encryption do not occur with correct input data,
	// but you can check that the function works correctly
	encrypted, err := EncryptData([]byte("test"), "password")
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
}

func TestDecryptData_Error(t *testing.T) {
	// test case when ciphertext is too short
	_, err := DecryptData([]byte("tooshort"), "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")

	// test case when ciphertext is damaged
	encrypted, _ := EncryptData([]byte("test"), "password")
	if len(encrypted) > 0 {
		encrypted[len(encrypted)-1] = encrypted[len(encrypted)-1] ^ 0xFF // invert the last byte
	}
	_, err = DecryptData(encrypted, "password")
	assert.Error(t, err)
}
