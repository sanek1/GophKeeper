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
			name:       "Простая строка",
			plaintext:  []byte("это секретное сообщение"),
			passphrase: "пароль123",
		},
		{
			name:       "Пустые данные",
			plaintext:  []byte(""),
			passphrase: "пароль123",
		},
		{
			name:       "Бинарные данные",
			plaintext:  []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			passphrase: "секретный_ключ",
		},
		{
			name:       "Длинный текст",
			plaintext:  bytes.Repeat([]byte("абвгдежзик"), 100),
			passphrase: "очень_длинный_пароль_для_тестирования",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Шифруем данные
			encrypted, err := EncryptData(tt.plaintext, tt.passphrase)
			assert.NoError(t, err)
			assert.NotNil(t, encrypted)
			assert.NotEqual(t, tt.plaintext, encrypted)

			// Расшифровываем данные
			decrypted, err := DecryptData(encrypted, tt.passphrase)
			assert.NoError(t, err)

			// Проверяем длину и содержимое
			if len(tt.plaintext) == 0 {
				// Для пустых данных проверяем только длину
				assert.Equal(t, 0, len(decrypted))
			} else {
				// Для непустых данных проверяем содержимое
				assert.Equal(t, tt.plaintext, decrypted)
			}

			// Проверяем с неверным паролем
			_, err = DecryptData(encrypted, tt.passphrase+"неверный")
			assert.Error(t, err)
		})
	}
}

func TestEncryptData_Error(t *testing.T) {
	// Ошибки при шифровании не возникают с правильными входными данными,
	// но можно проверить, что функция работает нормально
	encrypted, err := EncryptData([]byte("test"), "password")
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
}

func TestDecryptData_Error(t *testing.T) {
	// Тестируем случай, когда ciphertext слишком короткий
	_, err := DecryptData([]byte("tooshort"), "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")

	// Тестируем случай с поврежденными данными
	encrypted, _ := EncryptData([]byte("test"), "password")
	if len(encrypted) > 0 {
		encrypted[len(encrypted)-1] = encrypted[len(encrypted)-1] ^ 0xFF // инвертируем последний байт
	}
	_, err = DecryptData(encrypted, "password")
	assert.Error(t, err)
}
