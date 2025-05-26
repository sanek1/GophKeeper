package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sanek1/GophKeeper/internal/auth"
	"github.com/sanek1/GophKeeper/internal/crypto"
	"github.com/sanek1/GophKeeper/internal/models"
)

// Config contains client settings
type Config struct {
	ServerURL     string
	TokenFile     string
	MasterPwdFile string
	CacheDir      string
	SyncInterval  time.Duration
}

// Client represents a client for working with the GophKeeper API
type Client struct {
	config     Config
	token      string
	masterPwd  string // Master password for encryption/decryption of data
	localCache map[string]models.Secret
	cacheMutex sync.RWMutex
	lastSync   time.Time
	syncing    bool
	syncMutex  sync.Mutex
}

var testJWTSecret = "test-secret-key-for-ci"

// NewClient creates a new client instance
func NewClient(config Config) *Client {
	if config.TokenFile == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.TokenFile = filepath.Join(homeDir, ".gophkeeper", "token")
		} else {
			config.TokenFile = ".gophkeeper_token"
		}
	}

	if config.MasterPwdFile == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.MasterPwdFile = filepath.Join(homeDir, ".gophkeeper", "master.key")
		} else {
			config.MasterPwdFile = ".gophkeeper_master"
		}
	}

	if config.CacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.CacheDir = filepath.Join(homeDir, ".gophkeeper", "cache")
		} else {
			config.CacheDir = ".gophkeeper_cache"
		}
	}

	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute
	}

	// Create directories if they don't exist
	ensureDirExists(filepath.Dir(config.TokenFile))
	ensureDirExists(filepath.Dir(config.MasterPwdFile))
	ensureDirExists(config.CacheDir)

	client := &Client{
		config:     config,
		localCache: make(map[string]models.Secret),
	}

	// Try to load token and master password if they exist
	client.loadToken()
	client.loadMasterPassword()
	client.loadLocalCache()

	return client
}

// Register registers a new user
func (c *Client) Register(login, password string) error {
	reqBody, err := json.Marshal(models.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	fmt.Println(reqBody)
	resp, err := http.Post(
		fmt.Sprintf("%s/api/register", c.config.ServerURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error during registration: %s (code %d)", body, resp.StatusCode)
	}

	return nil
}

// Login performs user login and saves the token
func (c *Client) Login(login, password string) error {
	reqBody, err := json.Marshal(models.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/login", c.config.ServerURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error during login: %s (code %d)", body, resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Check that the token is not empty
	if result.Token == "" {
		return fmt.Errorf("server returned empty token")
	}

	// Save the token
	c.token = result.Token
	if err := c.saveToken(); err != nil {
		return err
	}

	// Check that the token works
	if err := c.TestAuthentication(); err != nil {
		// Use the auth package to regenerate the token
		newToken, regErr := auth.RegenerateToken(c.token, testJWTSecret	)
		if regErr != nil {
			return fmt.Errorf("error during token regeneration: %w", regErr)
		}

		// Save the new token
		c.token = newToken
		if err := c.saveToken(); err != nil {
			return err
		}

		// Check the new token
		if err := c.TestAuthentication(); err != nil {
			return fmt.Errorf("login completed, but token is invalid: %w\nPossible, the server uses a non-standard JWT_SECRET", err)
		}

		fmt.Println("Token successfully regenerated and passed the check!")
	}

	// Perform a safe sync on the first login
	if err := c.safeSyncAfterLogin(); err != nil {
		// If the sync failed, but the login was successful, only log
		fmt.Printf("Warning: Sync not performed: %v\n", err)
	}

	return nil
}

// safeSyncAfterLogin performs a safe sync after login,
// which does not clear the token in case of an error
func (c *Client) safeSyncAfterLogin() error {
	c.syncMutex.Lock()
	defer c.syncMutex.Unlock()

	// Prevent duplicate sync
	if c.syncing {
		return nil
	}

	c.syncing = true
	defer func() { c.syncing = false }()

	// Get all secrets from the server
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/secrets", c.config.ServerURL), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// On the first login, we do not clear the token even if the sync failed
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error during sync after login: %s (code %d)", body, resp.StatusCode)
	}

	var secrets []models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return err
	}

	// Update the local cache
	c.cacheMutex.Lock()
	c.localCache = make(map[string]models.Secret)
	for _, secret := range secrets {
		c.localCache[secret.ID.String()] = secret
	}
	c.cacheMutex.Unlock()

	// Save the updated cache
	if err := c.saveLocalCache(); err != nil {
		return err
	}

	c.lastSync = time.Now()

	return nil
}

// SetMasterPassword sets the master password for encryption/decryption
func (c *Client) SetMasterPassword(password string) error {
	// Check the password (minimum 8 characters)
	if len(password) < 8 {
		return fmt.Errorf("master password must contain at least 8 characters")
	}

	c.masterPwd = password
	return c.saveMasterPassword()
}

// EncryptData encrypts data using the master password
func (c *Client) EncryptData(data []byte) ([]byte, error) {
	if c.masterPwd == "" {
		return nil, fmt.Errorf("master password is not set, use set-master-password")
	}

	return crypto.EncryptData(data, c.masterPwd)
}

// DecryptData decrypts data using the master password
func (c *Client) DecryptData(data []byte) ([]byte, error) {
	if c.masterPwd == "" {
		return nil, fmt.Errorf("master password is not set, use set-master-password")
	}

	return crypto.DecryptData(data, c.masterPwd)
}

// GetSecrets gets a list of all user secrets
func (c *Client) GetSecrets() ([]models.Secret, error) {
	// Check if data needs to be synchronized
	if time.Since(c.lastSync) > c.config.SyncInterval {
		if err := c.SyncWithServer(); err != nil {
			// If there was a sync error, but we have local data,
			// simply issue a warning and return local data
			if len(c.localCache) > 0 {
				return c.GetOfflineSecrets(), fmt.Errorf("sync error: %w, using local data", err)
			}
			return nil, fmt.Errorf("sync error: %w", err)
		}
	}

	// Return data from the local cache
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	secrets := make([]models.Secret, 0, len(c.localCache))
	for _, secret := range c.localCache {
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// GetSecret gets a secret by its ID
func (c *Client) GetSecret(id string) (*models.Secret, error) {
	// First check the local cache
	c.cacheMutex.RLock()
	secret, found := c.localCache[id]
	c.cacheMutex.RUnlock()

	if found {
		return &secret, nil
	}

	// If not found in the cache, request from the server
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/secrets/%s", c.config.ServerURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle authorization error
	if resp.StatusCode == http.StatusUnauthorized {
		c.token = ""
		_ = c.saveToken()
		return nil, fmt.Errorf("authorization error: token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting secret: %s (code %d)", body, resp.StatusCode)
	}

	var result models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Add to cache
	c.cacheMutex.Lock()
	c.localCache[id] = result
	c.cacheMutex.Unlock()

	// Save the updated cache
	c.saveLocalCache()

	return &result, nil
}

// CreateSecret creates a new secret
func (c *Client) CreateSecret(secretType, metadata string, data []byte) (*models.Secret, error) {
	// Encrypt data before sending
	encryptedData, err := c.EncryptData(data)
	if err != nil {
		return nil, fmt.Errorf("error encrypting data: %w", err)
	}

	reqBody, err := json.Marshal(models.SecretRequest{
		Type:     secretType,
		Metadata: metadata,
		Data:     encryptedData,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/secrets", c.config.ServerURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle authorization error
	if resp.StatusCode == http.StatusUnauthorized {
		c.token = ""
		_ = c.saveToken()
		return nil, fmt.Errorf("authorization error: token is invalid or expired")
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error creating secret: %s (code %d)", body, resp.StatusCode)
	}

	var result models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Add to local cache
	c.cacheMutex.Lock()
	c.localCache[result.ID.String()] = result
	c.cacheMutex.Unlock()

	// Save the updated cache
	c.saveLocalCache()

	return &result, nil
}

// UpdateSecret updates an existing secret
func (c *Client) UpdateSecret(id, secretType, metadata string, data []byte) error {
	// Encrypt data before sending
	encryptedData, err := c.EncryptData(data)
	if err != nil {
		return fmt.Errorf("error encrypting data: %w", err)
	}

	reqBody, err := json.Marshal(models.SecretRequest{
		Type:     secretType,
		Metadata: metadata,
		Data:     encryptedData,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/secrets/%s", c.config.ServerURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle authorization error
	if resp.StatusCode == http.StatusUnauthorized {
		c.token = ""
		_ = c.saveToken()
		return fmt.Errorf("authorization error: token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error updating secret: %s (code %d)", body, resp.StatusCode)
	}

	var result models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Update the local cache
	c.cacheMutex.Lock()
	c.localCache[id] = result
	c.cacheMutex.Unlock()

	// Save the updated cache
	c.saveLocalCache()

	return nil
}

// DeleteSecret deletes a secret by its ID
func (c *Client) DeleteSecret(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/secrets/%s", c.config.ServerURL, id), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle authorization error
	if resp.StatusCode == http.StatusUnauthorized {
		c.token = ""
		_ = c.saveToken()
		return fmt.Errorf("authorization error: token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting secret: %s (code %d)", body, resp.StatusCode)
	}

	// Remove from local cache
	c.cacheMutex.Lock()
	delete(c.localCache, id)
	c.cacheMutex.Unlock()

	// Save the updated cache
	c.saveLocalCache()

	return nil
}

// SyncWithServer synchronizes local data with the server
func (c *Client) SyncWithServer() error {
	c.syncMutex.Lock()
	defer c.syncMutex.Unlock()

	// Check if there is a token
	if c.token == "" {
		return fmt.Errorf("synchronization is not possible: user is not authorized")
	}

	// Prevent duplicate synchronization
	if c.syncing {
		return nil
	}

	c.syncing = true
	defer func() { c.syncing = false }()

	// Get all secrets from the server
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/secrets", c.config.ServerURL), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{
		Timeout: 10 * time.Second, // Increase timeout for synchronization
	}
	resp, err := client.Do(req)
	if err != nil {
		// Connection error - probably the server is not available
		// Do not clear the token, return the error
		return fmt.Errorf("connection error to the server: %w", err)
	}
	defer resp.Body.Close()

	// Check the response code
	if resp.StatusCode == http.StatusUnauthorized {
		// Authorization problem - return a detailed message
		body, _ := io.ReadAll(resp.Body)
		c.token = ""
		_ = c.saveToken() // Ignore the error when saving an empty token

		// Possible problem with JWT-secret - give a hint
		return fmt.Errorf("authorization error during synchronization: %s\nPossible, you need to re-authorize or update the JWT-secret", body)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("synchronization error: %s (code %d)", body, resp.StatusCode)
	}

	var secrets []models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return fmt.Errorf("error decoding the response: %w", err)
	}

	// Update the local cache
	c.cacheMutex.Lock()
	c.localCache = make(map[string]models.Secret)
	for _, secret := range secrets {
		c.localCache[secret.ID.String()] = secret
	}
	c.cacheMutex.Unlock()

	// Save the updated cache
	if err := c.saveLocalCache(); err != nil {
		return fmt.Errorf("error saving the cache: %w", err)
	}

	c.lastSync = time.Now()
	return nil
}

// StartAutoSync starts automatic synchronization in the background
func (c *Client) StartAutoSync(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.config.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = c.SyncWithServer() // Ignore errors to not interrupt the cycle
			case <-ctx.Done():
				return
			}
		}
	}()
}

// DisplaySecretData decrypts and returns secret data in a convenient format
func (c *Client) DisplaySecretData(secret *models.Secret) (string, error) {
	// Decrypt the data
	decryptedData, err := c.DecryptData(secret.Data)
	if err != nil {
		return "", fmt.Errorf("error decrypting data: %w", err)
	}

	// Depending on the secret type, format the output
	switch secret.Type {
	case "password":
		return fmt.Sprintf("Password: %s", string(decryptedData)), nil
	case "card":
		return fmt.Sprintf("Card data: %s", string(decryptedData)), nil
	case "text":
		return string(decryptedData), nil
	case "note":
		return string(decryptedData), nil
	case "file":
		return fmt.Sprintf("File content (size: %d bytes)", len(decryptedData)), nil
	default:
		return fmt.Sprintf("Data (%d bytes)", len(decryptedData)), nil
	}
}

// loadToken loads the token from the file
func (c *Client) loadToken() {
	data, err := os.ReadFile(c.config.TokenFile)
	if err == nil && len(data) > 0 {
		c.token = string(data)
	}
}

// saveToken saves the token to the file
func (c *Client) saveToken() error {
	return os.WriteFile(c.config.TokenFile, []byte(c.token), 0600)
}

// loadMasterPassword loads the master password from the file
func (c *Client) loadMasterPassword() {
	data, err := os.ReadFile(c.config.MasterPwdFile)
	if err == nil && len(data) > 0 {
		c.masterPwd = string(data)
	}
}

// saveMasterPassword saves the master password to the file
func (c *Client) saveMasterPassword() error {
	return os.WriteFile(c.config.MasterPwdFile, []byte(c.masterPwd), 0600)
}

// loadLocalCache loads the local cache from the file
func (c *Client) loadLocalCache() {
	cacheFile := filepath.Join(c.config.CacheDir, "cache.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return
	}

	var secrets []models.Secret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return
	}

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.localCache = make(map[string]models.Secret)
	for _, secret := range secrets {
		c.localCache[secret.ID.String()] = secret
	}
}

// saveLocalCache saves the local cache to the file
func (c *Client) saveLocalCache() error {
	c.cacheMutex.RLock()
	secrets := make([]models.Secret, 0, len(c.localCache))
	for _, secret := range c.localCache {
		secrets = append(secrets, secret)
	}
	c.cacheMutex.RUnlock()

	data, err := json.Marshal(secrets)
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(c.config.CacheDir, "cache.json")
	return os.WriteFile(cacheFile, data, 0600)
}

// ensureDirExists creates a directory if it does not exist
func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
	}
}

// GetOfflineSecrets returns secrets from the local cache even without connecting to the server
func (c *Client) GetOfflineSecrets() []models.Secret {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	secrets := make([]models.Secret, 0, len(c.localCache))
	for _, secret := range c.localCache {
		secrets = append(secrets, secret)
	}

	return secrets
}

// IsAuthenticated checks if the user is authenticated
func (c *Client) IsAuthenticated() bool {
	return c.token != ""
}

// TestAuthentication checks the validity of the token by requesting an API
func (c *Client) TestAuthentication() error {
	if c.token == "" {
		return fmt.Errorf("no authorization token")
	}

	// Print debug information
	fmt.Println("Using JWT_SECRET:", testJWTSecret)
	fmt.Printf("Testing authentication with token: %s\n", c.token)

	// Make a simple request to the API that requires authorization
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/secrets", c.config.ServerURL), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{
		Timeout: 5 * time.Second, // Set a timeout to not wait too long
	}

	resp, err := client.Do(req)
	if err != nil {
		// If there is a connection error, do not clear the token - the server may be temporarily unavailable
		return fmt.Errorf("connection error to the server: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body for logging
	body, _ := io.ReadAll(resp.Body)

	// Print debug information
	fmt.Printf("Server response: %d - %s\n", resp.StatusCode, body)

	if resp.StatusCode == http.StatusUnauthorized {
		// The token is invalid, but do not clear it here
		// Token management should be handled in the calling code
		return fmt.Errorf("token is invalid or expired, server response: %s", body)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		// Any errors except authorization
		return fmt.Errorf("token check error, code: %d, response: %s", resp.StatusCode, body)
	}

	// Any response other than 401 means that the token was accepted by the server
	return nil
}

// IsMasterPasswordSet checks if the master password is set
func (c *Client) IsMasterPasswordSet() bool {
	return c.masterPwd != ""
}

// Logout performs user logout
func (c *Client) Logout() error {
	c.token = ""
	c.masterPwd = ""

	// Clear the cache
	c.cacheMutex.Lock()
	c.localCache = make(map[string]models.Secret)
	c.cacheMutex.Unlock()

	// Remove the token and master password files
	_ = os.Remove(c.config.TokenFile)
	_ = os.Remove(c.config.MasterPwdFile)

	return nil
}
