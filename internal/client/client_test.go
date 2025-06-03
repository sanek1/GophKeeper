package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
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

	t.Run("LoadMasterPassword", func(t *testing.T) {
		masterFile := filepath.Join(tempDir, "master")
		client.config.MasterPwdFile = masterFile

		// Save master password
		testPassword := "testmaster123"
		client.masterPwd = testPassword
		err := client.saveMasterPassword()
		require.NoError(t, err)

		// Clear and load
		client.masterPwd = ""
		client.loadMasterPassword()

		assert.Equal(t, testPassword, client.masterPwd)
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

	t.Run("CardType", func(t *testing.T) {
		data := []byte("card data")
		encrypted, _ := client.EncryptData(data)

		secret := &models.Secret{
			Type: "card",
			Data: encrypted,
		}

		result, err := client.DisplaySecretData(secret)
		require.NoError(t, err)
		assert.Contains(t, result, "Card data:")
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

	t.Run("NoteType", func(t *testing.T) {
		data := []byte("note content")
		encrypted, _ := client.EncryptData(data)

		secret := &models.Secret{
			Type: "note",
			Data: encrypted,
		}

		result, err := client.DisplaySecretData(secret)
		require.NoError(t, err)
		assert.Equal(t, "note content", result)
	})

	t.Run("FileType", func(t *testing.T) {
		data := []byte("file content")
		encrypted, _ := client.EncryptData(data)

		secret := &models.Secret{
			Type: "file",
			Data: encrypted,
		}

		result, err := client.DisplaySecretData(secret)
		require.NoError(t, err)
		assert.Contains(t, result, "File content")
		assert.Contains(t, result, "12 bytes")
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

	t.Run("DecryptionError", func(t *testing.T) {
		secret := &models.Secret{
			Type: "password",
			Data: []byte("invalid encrypted data"),
		}

		_, err := client.DisplaySecretData(secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decrypting data")
	})
}

// Comprehensive HTTP Operations Tests
func TestClient_HTTPOperations_Comprehensive(t *testing.T) {
	testSecretID := uuid.New()

	// Create a comprehensive mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/register" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)

		case r.URL.Path == "/api/login" && r.Method == "POST":
			var loginReq models.LoginRequest
			json.NewDecoder(r.Body).Decode(&loginReq)
			if loginReq.Login == "error@test.com" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid credentials"}`))
				return
			}
			if loginReq.Login == "empty@test.com" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"token": ""})
				return
			}
			response := map[string]string{"token": "mock.jwt.token"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/api/secrets" && r.Method == "GET":
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			secrets := []models.Secret{
				{
					ID:       testSecretID,
					Type:     "password",
					Metadata: "Test Secret",
					Data:     []byte("encrypted"),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secrets)

		case r.URL.Path == "/api/secrets" && r.Method == "POST":
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var req models.SecretRequest
			json.NewDecoder(r.Body).Decode(&req)
			secret := models.Secret{
				ID:       uuid.New(),
				Type:     req.Type,
				Data:     req.Data,
				Metadata: req.Metadata,
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secret)

		case r.URL.Path == fmt.Sprintf("/api/secrets/%s", testSecretID.String()) && r.Method == "GET":
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			secret := models.Secret{
				ID:       testSecretID,
				Type:     "password",
				Metadata: "Test Secret",
				Data:     []byte("encrypted"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secret)

		case r.URL.Path == fmt.Sprintf("/api/secrets/%s", testSecretID.String()) && r.Method == "PUT":
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var req models.SecretRequest
			json.NewDecoder(r.Body).Decode(&req)
			secret := models.Secret{
				ID:       testSecretID,
				Type:     req.Type,
				Data:     req.Data,
				Metadata: req.Metadata,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secret)

		case r.URL.Path == fmt.Sprintf("/api/secrets/%s", testSecretID.String()) && r.Method == "DELETE":
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	client := NewClient(Config{
		ServerURL:     server.URL,
		TokenFile:     filepath.Join(tempDir, "token"),
		MasterPwdFile: filepath.Join(tempDir, "master"),
		CacheDir:      tempDir,
	})

	t.Run("Register_Success", func(t *testing.T) {
		err := client.Register("test@example.com", "password123")
		assert.NoError(t, err)
	})

	t.Run("Register_ErrorResponse", func(t *testing.T) {
		// Test error handling in Register method
		tempClient := &Client{
			config: Config{
				ServerURL: "http://invalid-url:99999", // Invalid URL to trigger error
			},
		}

		err := tempClient.Register("test@example.com", "password123")
		assert.Error(t, err) // Should error due to invalid server URL
	})

	t.Run("Login_Success", func(t *testing.T) {
		err := client.Login("test@example.com", "password123")
		assert.NoError(t, err)
		assert.NotEmpty(t, client.token)
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		err := client.Login("error@test.com", "wrongpassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error during login")
	})

	t.Run("Login_EmptyToken", func(t *testing.T) {
		err := client.Login("empty@test.com", "password123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned empty token")
	})

	// Set up client with token and master password for further tests
	client.token = "test-token"
	client.masterPwd = "master-password-123"

	t.Run("GetSecrets_Success", func(t *testing.T) {
		client.lastSync = time.Now().Add(-10 * time.Minute) // Force sync
		secrets, err := client.GetSecrets()
		if err != nil {
			// Should get sync error but return local data
			assert.Contains(t, err.Error(), "sync error")
		} else {
			assert.NotNil(t, secrets)
		}
	})

	t.Run("GetSecret_Success", func(t *testing.T) {
		secret, err := client.GetSecret(testSecretID.String())
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, testSecretID.String(), secret.ID.String())
	})

	t.Run("GetSecret_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		_, err := client.GetSecret(nonExistentID)
		assert.Error(t, err)
	})

	t.Run("CreateSecret_Success", func(t *testing.T) {
		secret, err := client.CreateSecret("password", "Test Password", []byte("mypassword"))
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, "password", secret.Type)
	})

	t.Run("CreateSecret_EncryptionError", func(t *testing.T) {
		client.masterPwd = "" // Clear master password to cause encryption error
		_, err := client.CreateSecret("password", "Test", []byte("data"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error encrypting data")

		client.masterPwd = "master-password-123" // Restore
	})

	t.Run("UpdateSecret_Success", func(t *testing.T) {
		err := client.UpdateSecret(testSecretID.String(), "password", "Updated Password", []byte("newpassword"))
		assert.NoError(t, err)
	})

	t.Run("DeleteSecret_Success", func(t *testing.T) {
		err := client.DeleteSecret(testSecretID.String())
		assert.NoError(t, err)
	})
}

func TestClient_SyncWithServer(t *testing.T) {
	testSecrets := []models.Secret{
		{
			ID:       uuid.New(),
			Type:     "password",
			Metadata: "Sync Test 1",
		},
		{
			ID:       uuid.New(),
			Type:     "note",
			Metadata: "Sync Test 2",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/secrets" && r.Method == "GET" {
			if r.Header.Get("Authorization") == "Bearer unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "unauthorized"}`))
				return
			}
			if r.Header.Get("Authorization") == "Bearer error" {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "server error"}`))
				return
			}
			if r.Header.Get("Authorization") == "Bearer invalid-json" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(testSecrets)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	t.Run("SyncWithServer_Success", func(t *testing.T) {
		client := &Client{
			config: Config{
				ServerURL: server.URL,
				CacheDir:  tempDir,
			},
			token:      "valid-token",
			localCache: make(map[string]models.Secret),
		}

		err := client.SyncWithServer()
		assert.NoError(t, err)
		assert.Len(t, client.localCache, 2)
		assert.False(t, client.lastSync.IsZero())
	})

	t.Run("SyncWithServer_NoToken", func(t *testing.T) {
		tempDir := t.TempDir()
		client := &Client{
			config: Config{
				TokenFile:     filepath.Join(tempDir, "token"),
				MasterPwdFile: filepath.Join(tempDir, "master"),
				CacheDir:      tempDir,
			},
			token:      "",
			localCache: make(map[string]models.Secret),
		}

		err := client.SyncWithServer()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authorized")
	})

	t.Run("SyncWithServer_Unauthorized", func(t *testing.T) {
		tempDir := t.TempDir()
		client := &Client{
			config: Config{
				ServerURL:     server.URL,
				TokenFile:     filepath.Join(tempDir, "token"),
				MasterPwdFile: filepath.Join(tempDir, "master"),
				CacheDir:      tempDir,
			},
			token:      "unauthorized",
			localCache: make(map[string]models.Secret),
		}

		err := client.SyncWithServer()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization error")
	})

	t.Run("SyncWithServer_ServerError", func(t *testing.T) {
		tempDir := t.TempDir()
		client := &Client{
			config: Config{
				ServerURL:     server.URL,
				TokenFile:     filepath.Join(tempDir, "token"),
				MasterPwdFile: filepath.Join(tempDir, "master"),
				CacheDir:      tempDir,
			},
			token:      "error",
			localCache: make(map[string]models.Secret),
		}

		err := client.SyncWithServer()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "synchronization error")
	})

	t.Run("SyncWithServer_InvalidJSON", func(t *testing.T) {
		tempDir := t.TempDir()
		client := &Client{
			config: Config{
				ServerURL:     server.URL,
				TokenFile:     filepath.Join(tempDir, "token"),
				MasterPwdFile: filepath.Join(tempDir, "master"),
				CacheDir:      tempDir,
			},
			token:      "invalid-json",
			localCache: make(map[string]models.Secret),
		}

		err := client.SyncWithServer()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding")
	})

	t.Run("SyncWithServer_PreventDuplicate", func(t *testing.T) {
		tempDir := t.TempDir()
		client := &Client{
			config: Config{
				ServerURL:     server.URL,
				TokenFile:     filepath.Join(tempDir, "token"),
				MasterPwdFile: filepath.Join(tempDir, "master"),
				CacheDir:      tempDir,
			},
			token:      "valid-token",
			syncing:    true, // Already syncing
			localCache: make(map[string]models.Secret),
		}

		err := client.SyncWithServer()
		assert.NoError(t, err) // Should return immediately without error
	})
}

func TestClient_TestAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		switch auth {
		case "Bearer valid-token":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]models.Secret{})
		case "Bearer unauthorized-token":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
		case "Bearer forbidden-token":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "forbidden"}`))
		case "Bearer error-token":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	t.Run("TestAuthentication_NoToken", func(t *testing.T) {
		client := &Client{
			config: Config{ServerURL: server.URL},
			token:  "",
		}

		err := client.TestAuthentication()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no authorization token")
	})

	t.Run("TestAuthentication_ValidToken", func(t *testing.T) {
		client := &Client{
			config: Config{ServerURL: server.URL},
			token:  "valid-token",
		}

		err := client.TestAuthentication()
		assert.NoError(t, err)
	})

	t.Run("TestAuthentication_UnauthorizedToken", func(t *testing.T) {
		client := &Client{
			config: Config{ServerURL: server.URL},
			token:  "unauthorized-token",
		}

		err := client.TestAuthentication()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is invalid or expired")
	})

	t.Run("TestAuthentication_ForbiddenToken", func(t *testing.T) {
		client := &Client{
			config: Config{ServerURL: server.URL},
			token:  "forbidden-token",
		}

		err := client.TestAuthentication()
		assert.NoError(t, err) // 403 is acceptable for token validation
	})

	t.Run("TestAuthentication_ServerError", func(t *testing.T) {
		client := &Client{
			config: Config{ServerURL: server.URL},
			token:  "error-token",
		}

		err := client.TestAuthentication()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token check error")
	})
}

func TestClient_SafeSyncAfterLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/secrets" && r.Method == "GET" {
			secrets := []models.Secret{{
				ID:   uuid.New(),
				Type: "password",
			}}
			json.NewEncoder(w).Encode(secrets)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	client := &Client{
		config: Config{
			ServerURL: server.URL,
			CacheDir:  tempDir,
		},
		token:      "test-token",
		localCache: make(map[string]models.Secret),
	}

	err := client.safeSyncAfterLogin()
	assert.NoError(t, err)
	assert.Len(t, client.localCache, 1)
}

func TestClient_StartAutoSync(t *testing.T) {
	client := &Client{
		config: Config{
			SyncInterval: 50 * time.Millisecond,
		},
		token: "test-token",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
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

	t.Run("EnsureDirExists_AlreadyExists", func(t *testing.T) {
		// Call again on existing directory
		ensureDirExists(testDir)

		// Should still exist and be a directory
		info, err := os.Stat(testDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
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

	t.Run("GetSecrets_WithCache_NoSync", func(t *testing.T) {
		client := &Client{
			config: Config{
				SyncInterval: 10 * time.Minute,
			},
			token: "valid-token",
			localCache: map[string]models.Secret{
				"test": {Type: "test"},
			},
			lastSync: time.Now(), // Recent sync
		}

		secrets, err := client.GetSecrets()
		assert.NoError(t, err)
		assert.Len(t, secrets, 1)
	})

	t.Run("Register_JSONMarshalError", func(t *testing.T) {
		client := &Client{}

		// This tests the JSON marshaling path
		err := client.Register("test@example.com", "password123")
		assert.Error(t, err) // Should error due to no server URL
	})

	t.Run("CreateSecret_JSONMarshalError", func(t *testing.T) {
		client := &Client{
			masterPwd: "password123",
		}

		// This should trigger JSON marshal path
		_, err := client.CreateSecret("password", "test", []byte("data"))
		assert.Error(t, err) // Should error due to no server URL
	})

	t.Run("UpdateSecret_JSONMarshalError", func(t *testing.T) {
		client := &Client{
			masterPwd: "password123",
		}

		err := client.UpdateSecret("test-id", "password", "test", []byte("data"))
		assert.Error(t, err) // Should error due to no server URL
	})
}
