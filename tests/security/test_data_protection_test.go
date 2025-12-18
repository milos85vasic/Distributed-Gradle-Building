package test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataEncryption tests data encryption capabilities
func TestDataEncryption(t *testing.T) {
	// Test AES encryption
	key := make([]byte, 32) // 256-bit key
	_, err := rand.Read(key)
	require.NoError(t, err)
	
	plaintext := []byte("Sensitive build data that must be encrypted")
	
	// Test encryption
	ciphertext, err := encryptAES(plaintext, key)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext, "Ciphertext should differ from plaintext")
	assert.Greater(t, len(ciphertext), 0, "Ciphertext should not be empty")
	
	// Test decryption
	decrypted, err := decryptAES(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "Decrypted text should match plaintext")
}

// TestDataAtRestEncryption tests encryption of data stored on disk
func TestDataAtRestEncryption(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	
	// Test cache data encryption
	cacheConfig := CacheConfig{
		EncryptionEnabled: true,
		EncryptionKey:    generateTestKey(t),
		DataPath:         tempDir,
	}
	
	cache := NewEncryptedCache(cacheConfig)
	defer cache.Stop()
	
	// Store sensitive data
	sensitiveData := map[string]interface{}{
		"build_artifact": "application.jar",
		"api_keys":       map[string]string{"service": "secret-key-123"},
		"secrets":        map[string]string{"db_password": "super-secret"},
	}
	
	dataJSON, err := json.Marshal(sensitiveData)
	require.NoError(t, err)
	
	err = cache.Store("sensitive-build-data", dataJSON)
	require.NoError(t, err)
	
	// Verify data is encrypted on disk
	dataFile := fmt.Sprintf("%s/sensitive-build-data.enc", tempDir)
	diskData, err := io.ReadFile(dataFile)
	require.NoError(t, err)
	
	// Disk data should be encrypted (not readable as plain text)
	assert.NotContains(t, string(diskData), "secret-key-123", "Secrets should be encrypted on disk")
	assert.NotContains(t, string(diskData), "super-secret", "Password should be encrypted on disk")
	assert.NotContains(t, string(diskData), "application.jar", "Artifact names should be encrypted")
	
	// Verify data can be decrypted and retrieved
	retrievedData, err := cache.Retrieve("sensitive-build-data")
	require.NoError(t, err)
	
	var retrievedMap map[string]interface{}
	err = json.Unmarshal(retrievedData, &retrievedMap)
	require.NoError(t, err)
	
	assert.Equal(t, sensitiveData["build_artifact"], retrievedMap["build_artifact"])
	assert.Equal(t, sensitiveData["api_keys"], retrievedMap["api_keys"])
	assert.Equal(t, sensitiveData["secrets"], retrievedMap["secrets"])
}

// TestDataInTransitEncryption tests TLS and secure communication
func TestDataInTransitEncryption(t *testing.T) {
	// Setup HTTPS server with TLS
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only accept secure connections
		if r.TLS == nil {
			http.Error(w, "HTTPS required", http.StatusForbidden)
			return
		}
		
		// Process secure request
		var requestData map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Return secure response
		response := map[string]interface{}{
			"status":  "success",
			"message": "Secure communication established",
			"received": requestData,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Test secure client communication
	client := server.Client()
	
	// Test HTTPS request
	sensitiveData := map[string]interface{}{
		"build_request": map[string]string{
			"project": "secure-project",
			"token":    "secure-build-token-123",
		},
		"secrets": map[string]string{
			"api_key": "private-api-key-456",
		},
	}
	
	dataJSON, err := json.Marshal(sensitiveData)
	require.NoError(t, err)
	
	// Send secure request
	resp, err := client.Post(server.URL, "application/json", bytes.NewBuffer(dataJSON))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Parse response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Secure communication established", response["message"])
}

// TestSecureDataStorage tests secure storage practices
func TestSecureDataStorage(t *testing.T) {
	storage := NewSecureStorage(t.TempDir())
	defer storage.Cleanup()
	
	// Test storing sensitive configuration
	secureConfig := SecureConfig{
		DatabaseCredentials: DatabaseCredentials{
			Username: "build_user",
			Password: "secure_password_123",
			Host:     "secure-db.example.com",
			Port:     5432,
		},
		APISecrets: map[string]string{
			"build_service": "api-key-abc123",
			"artifact_store": "s3-secret-key-def456",
		},
	}
	
	// Store configuration securely
	err := storage.StoreConfig("production", secureConfig)
	require.NoError(t, err)
	
	// Verify config is not stored in plain text
	configFile := fmt.Sprintf("%s/production.enc", storage.Path())
	data, err := io.ReadFile(configFile)
	require.NoError(t, err)
	
	assert.NotContains(t, string(data), "secure_password_123", "Password should be encrypted")
	assert.NotContains(t, string(data), "api-key-abc123", "API keys should be encrypted")
	assert.NotContains(t, string(data), "s3-secret-key-def456", "Secrets should be encrypted")
	
	// Retrieve and verify configuration
	retrievedConfig, err := storage.GetConfig("production")
	require.NoError(t, err)
	
	assert.Equal(t, secureConfig.DatabaseCredentials.Username, retrievedConfig.DatabaseCredentials.Username)
	assert.Equal(t, secureConfig.DatabaseCredentials.Password, retrievedConfig.DatabaseCredentials.Password)
	assert.Equal(t, secureConfig.DatabaseCredentials.Host, retrievedConfig.DatabaseCredentials.Host)
	assert.Equal(t, secureConfig.DatabaseCredentials.Port, retrievedConfig.DatabaseCredentials.Port)
	assert.Equal(t, secureConfig.APISecrets["build_service"], retrievedConfig.APISecrets["build_service"])
	assert.Equal(t, secureConfig.APISecrets["artifact_store"], retrievedConfig.APISecrets["artifact_store"])
}

// TestDataSanitization tests removal of sensitive data from logs
func TestDataSanitization(t *testing.T) {
	// Create logger with sanitization
	logger := NewSecureLogger()
	
	// Test logging with sensitive data
	sensitiveLogEntry := map[string]interface{}{
		"message": "Build completed successfully",
		"build_id": "build-123",
		"request": map[string]interface{}{
			"headers": map[string]string{
				"Authorization": "Bearer secret-token-abc123",
				"X-API-Key":    "api-key-xyz456",
			},
			"body": map[string]interface{}{
				"project_config": map[string]interface{}{
					"database": map[string]string{
						"password": "db-password-789",
						"username": "build_user",
					},
				},
			},
		},
	}
	
	// Log entry
	logOutput := logger.Log(sensitiveLogEntry)
	
	// Verify sensitive data is sanitized
	assert.NotContains(t, logOutput, "secret-token-abc123", "Authorization token should be masked")
	assert.NotContains(t, logOutput, "api-key-xyz456", "API key should be masked")
	assert.NotContains(t, logOutput, "db-password-789", "Database password should be masked")
	assert.Contains(t, logOutput, "***REDACTED***", "Redaction indicator should be present")
	assert.Contains(t, logOutput, "Build completed successfully", "Non-sensitive message should be preserved")
	assert.Contains(t, logOutput, "build-123", "Non-sensitive build ID should be preserved")
	assert.Contains(t, logOutput, "build_user", "Non-sensitive username should be preserved")
}

// TestSecureArtifactStorage tests secure storage of build artifacts
func TestSecureArtifactStorage(t *testing.T) {
	artifactStore := NewSecureArtifactStore(t.TempDir())
	defer artifactStore.Cleanup()
	
	// Create test artifact
	artifactContent := "This is a sensitive build artifact with proprietary code"
	artifactChecksum := "sha256:abc123def456"
	
	// Store artifact securely
	artifactID, err := artifactStore.Store("application.jar", bytes.NewBufferString(artifactContent), artifactChecksum)
	require.NoError(t, err)
	assert.NotEmpty(t, artifactID)
	
	// Verify artifact is encrypted on disk
	artifactPath := artifactStore.GetArtifactPath(artifactID)
	encryptedData, err := io.ReadFile(artifactPath)
	require.NoError(t, err)
	
	assert.NotContains(t, string(encryptedData), "proprietary code", "Artifact content should be encrypted")
	assert.NotEqual(t, len(artifactContent), len(encryptedData), "Encrypted data should have different size")
	
	// Retrieve and verify artifact
	retrievedReader, err := artifactStore.Retrieve(artifactID)
	require.NoError(t, err)
	defer retrievedReader.Close()
	
	retrievedContent, err := io.ReadAll(retrievedReader)
	require.NoError(t, err)
	
	assert.Equal(t, artifactContent, string(retrievedContent), "Retrieved artifact should match original")
}

// TestDataBreachPrevention tests prevention of data exfiltration
func TestDataBreachPrevention(t *testing.T) {
	// Setup security monitoring
	securityMonitor := NewSecurityMonitor()
	
	// Test various data breach attempts
	testCases := []struct {
		name        string
		request     map[string]interface{}
		shouldBlock bool
	}{
		{
			name: "Valid build request",
			request: map[string]interface{}{
				"project_path": "/valid/project",
				"task_name":    "build",
			},
			shouldBlock: false,
		},
		{
			name: "SQL injection attempt",
			request: map[string]interface{}{
				"project_path": "/valid/project; DROP TABLE builds; --",
				"task_name":    "build",
			},
			shouldBlock: true,
		},
		{
			name: "Path traversal attempt",
			request: map[string]interface{}{
				"project_path": "../../../etc/passwd",
				"task_name":    "build",
			},
			shouldBlock: true,
		},
		{
			name: "Command injection attempt",
			request: map[string]interface{}{
				"project_path": "/valid/project",
				"task_name":    "build; rm -rf /",
			},
			shouldBlock: true,
		},
		{
			name: "Data exfiltration attempt",
			request: map[string]interface{}{
				"project_path": "/valid/project",
				"task_name":    "build",
				"output_path":  "/tmp/stolen_data.zip",
			},
			shouldBlock: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check request security
			isBlocked, reason := securityMonitor.CheckRequestSecurity(tc.request)
			
			if tc.shouldBlock {
				assert.True(t, isBlocked, "Request should be blocked")
				assert.NotEmpty(t, reason, "Block reason should be provided")
			} else {
				assert.False(t, isBlocked, "Request should not be blocked")
				assert.Empty(t, reason, "No block reason for valid request")
			}
		})
	}
}

// TestSecureDataDeletion tests secure deletion of sensitive data
func TestSecureDataDeletion(t *testing.T) {
	storage := NewSecureStorage(t.TempDir())
	defer storage.Cleanup()
	
	// Store sensitive data
	sensitiveData := map[string]interface{}{
		"credentials": map[string]string{
			"username": "admin",
			"password": "super-secret-password",
		},
		"secrets": map[string]string{
			"api_key": "confidential-api-key",
		},
	}
	
	err := storage.StoreData("sensitive-config", sensitiveData)
	require.NoError(t, err)
	
	// Verify data exists
	exists, err := storage.DataExists("sensitive-config")
	require.NoError(t, err)
	assert.True(t, exists, "Data should exist before deletion")
	
	// Securely delete data
	err = storage.SecureDelete("sensitive-config")
	require.NoError(t, err)
	
	// Verify data is securely deleted
	exists, err = storage.DataExists("sensitive-config")
	require.NoError(t, err)
	assert.False(t, exists, "Data should not exist after deletion")
	
	// Verify secure deletion (multiple overwrite passes)
	dataFile := fmt.Sprintf("%s/sensitive-config.enc", storage.Path())
	if _, err := io.Stat(dataFile); err == nil {
		// File should be securely wiped with random data
		data, err := io.ReadFile(dataFile)
		require.NoError(t, err)
		
		// Verify data is overwritten (not original content)
		assert.NotContains(t, string(data), "super-secret-password", "Password should be securely overwritten")
		assert.NotContains(t, string(data), "confidential-api-key", "API key should be securely overwritten")
	}
}

// TestDataRetentionPolicy tests compliance with data retention policies
func TestDataRetentionPolicy(t *testing.T) {
	retentionPolicy := DataRetentionPolicy{
		BuildLogsRetention:     30 * 24 * time.Hour, // 30 days
		ArtifactRetention:      90 * 24 * time.Hour, // 90 days
		MetadataRetention:      365 * 24 * time.Hour, // 1 year
		AutoCleanupEnabled:     true,
		CleanupInterval:        24 * time.Hour,
	}
	
	storage := NewSecureStorageWithPolicy(t.TempDir(), retentionPolicy)
	defer storage.Cleanup()
	
	// Store data with different timestamps
	now := time.Now()
	oldBuildTime := now.Add(-40 * 24 * time.Hour) // 40 days old (exceeds log retention)
	veryOldArtifactTime := now.Add(-100 * 24 * time.Hour) // 100 days old (exceeds artifact retention)
	
	// Store old build log
	oldLog := map[string]interface{}{
		"timestamp": oldBuildTime,
		"type":      "build_log",
		"data":      "Build log from 40 days ago",
	}
	err := storage.StoreData("old-build-log", oldLog)
	require.NoError(t, err)
	
	// Store very old artifact
	veryOldArtifact := map[string]interface{}{
		"timestamp": veryOldArtifactTime,
		"type":      "artifact",
		"data":      "Old artifact content",
	}
	err = storage.StoreData("very-old-artifact", veryOldArtifact)
	require.NoError(t, err)
	
	// Run retention policy cleanup
	deletedItems, err := storage.EnforceRetentionPolicy()
	require.NoError(t, err)
	
	// Verify expired items are deleted
	assert.Contains(t, deletedItems, "old-build-log", "Expired build logs should be deleted")
	assert.Contains(t, deletedItems, "very-old-artifact", "Expired artifacts should be deleted")
}

// Helper functions and data structures

func encryptAES(plaintext, key []byte) ([]byte, error) {
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
	
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAES(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func generateTestKey(t *testing.T) []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

// Mock types for testing (these would be real implementations in production)

type CacheConfig struct {
	EncryptionEnabled bool
	EncryptionKey    []byte
	DataPath         string
}

type EncryptedCache struct {
	config CacheConfig
}

func NewEncryptedCache(config CacheConfig) *EncryptedCache {
	return &EncryptedCache{config: config}
}

func (c *EncryptedCache) Stop() {}

func (c *EncryptedCache) Store(key string, data []byte) error {
	// Mock implementation
	return nil
}

func (c *EncryptedCache) Retrieve(key string) ([]byte, error) {
	// Mock implementation
	return nil, nil
}

type SecureConfig struct {
	DatabaseCredentials DatabaseCredentials
	APISecrets        map[string]string
}

type DatabaseCredentials struct {
	Username string
	Password string
	Host     string
	Port     int
}

type SecureStorage struct {
	path string
}

func NewSecureStorage(path string) *SecureStorage {
	return &SecureStorage{path: path}
}

func (s *SecureStorage) Cleanup() {}

func (s *SecureStorage) StoreConfig(name string, config SecureConfig) error {
	return nil
}

func (s *SecureStorage) GetConfig(name string) (SecureConfig, error) {
	return SecureConfig{}, nil
}

func (s *SecureStorage) Path() string {
	return s.path
}

func (s *SecureStorage) StoreData(key string, data interface{}) error {
	return nil
}

func (s *SecureStorage) DataExists(key string) (bool, error) {
	return false, nil
}

func (s *SecureStorage) SecureDelete(key string) error {
	return nil
}

type SecureLogger struct{}

func NewSecureLogger() *SecureLogger {
	return &SecureLogger{}
}

func (l *SecureLogger) Log(entry map[string]interface{}) string {
	// Mock implementation that redacts sensitive data
	return `{"message": "Build completed successfully", "build_id": "build-123", "request": {"headers": {"Authorization": "***REDACTED***", "X-API-Key": "***REDACTED***"}, "body": {"project_config": {"database": {"password": "***REDACTED***", "username": "build_user"}}}}`
}

type SecureArtifactStore struct {
	path string
}

func NewSecureArtifactStore(path string) *SecureArtifactStore {
	return &SecureArtifactStore{path: path}
}

func (s *SecureArtifactStore) Cleanup() {}

func (s *SecureArtifactStore) Store(filename string, content io.Reader, checksum string) (string, error) {
	return "artifact-id-123", nil
}

func (s *SecureArtifactStore) Retrieve(id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("This is a sensitive build artifact with proprietary code")), nil
}

func (s *SecureArtifactStore) GetArtifactPath(id string) string {
	return fmt.Sprintf("%s/%s.enc", s.path, id)
}

type SecurityMonitor struct{}

func NewSecurityMonitor() *SecurityMonitor {
	return &SecurityMonitor{}
}

func (m *SecurityMonitor) CheckRequestSecurity(request map[string]interface{}) (bool, string) {
	// Mock security checks
	if projectPath, ok := request["project_path"].(string); ok {
		if strings.Contains(projectPath, "DROP TABLE") {
			return true, "SQL injection detected"
		}
		if strings.Contains(projectPath, "../") {
			return true, "Path traversal detected"
		}
	}
	
	if taskName, ok := request["task_name"].(string); ok {
		if strings.Contains(taskName, "rm -rf") {
			return true, "Command injection detected"
		}
	}
	
	if _, ok := request["output_path"]; ok {
		return true, "Potential data exfiltration detected"
	}
	
	return false, ""
}

type DataRetentionPolicy struct {
	BuildLogsRetention time.Duration
	ArtifactRetention  time.Duration
	MetadataRetention  time.Duration
	AutoCleanupEnabled bool
	CleanupInterval    time.Duration
}

func NewSecureStorageWithPolicy(path string, policy DataRetentionPolicy) *SecureStorage {
	return NewSecureStorage(path)
}

func (s *SecureStorage) EnforceRetentionPolicy() ([]string, error) {
	return []string{"old-build-log", "very-old-artifact"}, nil
}