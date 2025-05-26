package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sanek1/GophKeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := Config{
			ServerURL: "http://localhost:8080",
		}

		client := NewClient(config)

		assert.NotNil(t, client)
		assert.Equal(t, "http://localhost:8080", client.config.ServerURL)
		assert.NotEmpty(t, client.config.TokenFile)
		assert.NotEmpty(t, client.config.MasterPwdFile)
		assert.NotEmpty(t, client.config.CacheDir)
		assert.Equal(t, 5*time.Minute, client.config.SyncInterval)
		assert.NotNil(t, client.localCache)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		tempDir := t.TempDir()
		config := Config{
			ServerURL:     "http://test:9000",
			TokenFile:     filepath.Join(tempDir, "token"),
			MasterPwdFile: filepath.Join(tempDir, "master"),
			CacheDir:      filepath.Join(tempDir, "cache"),
			SyncInterval:  10 * time.Minute,
		}

		client := NewClient(config)

		assert.Equal(t, config.ServerURL, client.config.ServerURL)
		assert.Equal(t, config.TokenFile, client.config.TokenFile)
		assert.Equal(t, config.SyncInterval, client.config.SyncInterval)
	})
}

func TestClient_SetMasterPassword(t *testing.T) {
	client := &Client{
		config: Config{
			MasterPwdFile: filepath.Join(t.TempDir(), "master"),
		},
	}

	t.Run("ValidPassword", func(t *testing.T) {
		err := client.SetMasterPassword("password123")
		assert.NoError(t, err)
		assert.Equal(t, "password123", client.masterPwd)
	})

	t.Run("TooShortPassword", func(t *testing.T) {
		err := client.SetMasterPassword("short")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters")
	})
}

func TestClient_EncryptDecryptData(t *testing.T) {
	client := &Client{masterPwd: "testpassword123"}

	t.Run("SuccessfulEncryptDecrypt", func(t *testing.T) {
		originalData := []byte("test secret data")

		encrypted, err := client.EncryptData(originalData)
		require.NoError(t, err)
		assert.NotEqual(t, originalData, encrypted)

		decrypted, err := client.DecryptData(encrypted)
		require.NoError(t, err)
		assert.Equal(t, originalData, decrypted)
	})

	t.Run("NoMasterPassword", func(t *testing.T) {
		clientNoPassword := &Client{}

		_, err := clientNoPassword.EncryptData([]byte("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "master password is not set")

		_, err = clientNoPassword.DecryptData([]byte("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "master password is not set")
	})
}

func TestClient_TokenOperations(t *testing.T) {
	tempDir := t.TempDir()
	client := &Client{
		config: Config{
			TokenFile: filepath.Join(tempDir, "token"),
		},
	}

	t.Run("SaveAndLoadToken", func(t *testing.T) {
		testToken := "test.jwt.token"
		client.token = testToken

		err := client.saveToken()
		require.NoError(t, err)

		// Clear the token and load it again
		client.token = ""
		client.loadToken()

		assert.Equal(t, testToken, client.token)
	})
}
func TestClient_CacheOperations(t *testing.T) {
	tempDir := t.TempDir()
	client := &Client{
		config: Config{
			CacheDir: tempDir,
		},
		localCache: make(map[string]models.Secret),
	}

	t.Run("SaveAndLoadCache", func(t *testing.T) {
		// Create test secrets
		secret1 := models.Secret{
			ID:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			Type:     "password",
			Data:     []byte("secret1"),
			Metadata: "Test Secret 1",
		}
		secret2 := models.Secret{
			ID:       [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			Type:     "note",
			Data:     []byte("secret2"),
			Metadata: "Test Secret 2",
		}

		client.localCache[secret1.ID.String()] = secret1
		client.localCache[secret2.ID.String()] = secret2

		err := client.saveLocalCache()
		require.NoError(t, err)

		// Clear the cache and load it again
		client.localCache = make(map[string]models.Secret)
		client.loadLocalCache()

		assert.Len(t, client.localCache, 2)
		assert.Equal(t, secret1, client.localCache[secret1.ID.String()])
		assert.Equal(t, secret2, client.localCache[secret2.ID.String()])
	})
}

func TestClient_IsAuthenticated(t *testing.T) {
	t.Run("WithToken", func(t *testing.T) {
		client := &Client{token: "some.jwt.token"}
		assert.True(t, client.IsAuthenticated())
	})

	t.Run("WithoutToken", func(t *testing.T) {
		client := &Client{token: ""}
		assert.False(t, client.IsAuthenticated())
	})
}

func TestClient_IsMasterPasswordSet(t *testing.T) {
	t.Run("WithPassword", func(t *testing.T) {
		client := &Client{masterPwd: "password123"}
		assert.True(t, client.IsMasterPasswordSet())
	})

	t.Run("WithoutPassword", func(t *testing.T) {
		client := &Client{masterPwd: ""}
		assert.False(t, client.IsMasterPasswordSet())
	})
}

func TestClient_GetOfflineSecrets(t *testing.T) {
	client := &Client{
		localCache: make(map[string]models.Secret),
	}

	secret1 := models.Secret{
		ID:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Type:     "password",
		Metadata: "Test Secret 1",
	}
	secret2 := models.Secret{
		ID:       [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Type:     "note",
		Metadata: "Test Secret 2",
	}

	client.localCache[secret1.ID.String()] = secret1
	client.localCache[secret2.ID.String()] = secret2

	secrets := client.GetOfflineSecrets()
	assert.Len(t, secrets, 2)
}

func TestClient_DisplaySecretData(t *testing.T) {
	client := &Client{masterPwd: "testpassword123"}

	t.Run("PasswordType", func(t *testing.T) {
		data := []byte("mypassword")
		encrypted, _ := client.EncryptData(data)

		secret := &models.Secret{
			Type: "password",
			Data: encrypted,
		}

		result, err := client.DisplaySecretData(secret)
		require.NoError(t, err)
		assert.Contains(t, result, "Password:")
		assert.Contains(t, result, "mypassword")
	})

	t.Run("TextType", func(t *testing.T) {
		data := []byte("some text")
		encrypted, _ := client.EncryptData(data)

		secret := &models.Secret{
			Type: "text",
			Data: encrypted,
		}

		result, err := client.DisplaySecretData(secret)
		require.NoError(t, err)
		assert.Equal(t, "some text", result)
	})

	t.Run("UnknownType", func(t *testing.T) {
		data := []byte("unknown data")
		encrypted, _ := client.EncryptData(data)

		secret := &models.Secret{
			Type: "unknown",
			Data: encrypted,
		}

		result, err := client.DisplaySecretData(secret)
		require.NoError(t, err)
		assert.Contains(t, result, "Data (12 bytes)")
	})
}

func TestClient_HTTPOperations(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/register":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				return
			}
		case "/api/login":
			if r.Method == "POST" {
				response := map[string]string{"token": "mock.jwt.token"}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		case "/api/secrets":
			if r.Method == "GET" {
				// Mock response for getting secrets
				secrets := []models.Secret{
					{
						ID:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
						Type:     "password",
						Metadata: "Test",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(secrets)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	client := NewClient(Config{
		ServerURL:     server.URL,
		TokenFile:     filepath.Join(tempDir, "token"),
		MasterPwdFile: filepath.Join(tempDir, "master"),
		CacheDir:      tempDir,
	})

	t.Run("Register", func(t *testing.T) {
		err := client.Register("test@example.com", "password123")
		assert.NoError(t, err)
	})

	// For testing other methods, we need a more complex mock server
	// but this is a basic example
}

func TestClient_Logout(t *testing.T) {
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "token")
	masterFile := filepath.Join(tempDir, "master")

	// Create files
	err := os.WriteFile(tokenFile, []byte("token"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(masterFile, []byte("master"), 0600)
	require.NoError(t, err)

	client := &Client{
		config: Config{
			TokenFile:     tokenFile,
			MasterPwdFile: masterFile,
		},
		token:     "test-token",
		masterPwd: "test-password",
		localCache: map[string]models.Secret{
			"test": {Type: "test"},
		},
	}

	err = client.Logout()
	require.NoError(t, err)

	assert.Empty(t, client.token)
	assert.Empty(t, client.masterPwd)
	assert.Empty(t, client.localCache)

	// Check that the files are deleted
	_, err = os.Stat(tokenFile)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(masterFile)
	assert.True(t, os.IsNotExist(err))
}

func TestClient_StartAutoSync(t *testing.T) {
	client := &Client{
		config: Config{
			SyncInterval: 100 * time.Millisecond,
		},
		token: "test-token",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start auto synchronization
	client.StartAutoSync(ctx)

	// Wait for the context to complete
	<-ctx.Done()

	// The test checks that the function does not panic and correctly completes
}

func TestEnsureDirExists(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "dir")

	// The directory should not exist
	_, err := os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))

	// Create the directory
	ensureDirExists(testDir)

	// Check that the directory is created
	info, err := os.Stat(testDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// Additional tests for error cases and edge cases
func TestClient_ErrorCases(t *testing.T) {
	t.Run("SyncWithServer_NoToken", func(t *testing.T) {
		client := &Client{token: ""}
		err := client.SyncWithServer()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authorized")
	})

	t.Run("GetSecrets_EmptyCache", func(t *testing.T) {
		client := &Client{
			localCache: make(map[string]models.Secret),
			lastSync:   time.Now().Add(-10 * time.Minute), 
		}

		_, err := client.GetSecrets()
		assert.Error(t, err)
	})
}
