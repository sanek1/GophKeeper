package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// EncryptData encrypts data using AES-GCM
func EncryptData(plaintext []byte, passphrase string) ([]byte, error) {
	key := generateKey(passphrase)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// encryption of data and adding nonce to the beginning
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// DecryptData decrypts data encrypted using AES-GCM
func DecryptData(ciphertext []byte, passphrase string) ([]byte, error) {
	key := generateKey(passphrase)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// extract nonce from the beginning of the encrypted data
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	// decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// generateKey generates a 32-byte key from the passphrase
func generateKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}
