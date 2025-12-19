package cachepkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"distributed-gradle-building/types"
)

func TestNewCacheServer(t *testing.T) {
	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      "/tmp/test-cache",
		MaxCacheSize:    100 * 1024 * 1024, // 100MB
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)
	if server == nil {
		t.Fatal("Expected cache server to be created")
	}

	if server.Config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.Config.Port)
	}

	if server.Config.StorageType != "filesystem" {
		t.Errorf("Expected filesystem storage, got %s", server.Config.StorageType)
	}
}

func TestFileSystemStorage_GetPutDelete(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewFileSystemStorage(tempDir, time.Hour)

	// Test Put
	entry := &CacheEntry{
		Key:       "test-key",
		Data:      []byte("test data"),
		Timestamp: time.Now(),
		TTL:       time.Hour,
		Metadata:  map[string]string{"project": "test"},
	}

	err = storage.Put("test-key", entry)
	if err != nil {
		t.Fatalf("Failed to put cache entry: %v", err)
	}

	// Test Get
	retrieved, err := storage.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get cache entry: %v", err)
	}

	if string(retrieved.Data) != "test data" {
		t.Errorf("Expected data 'test data', got '%s'", string(retrieved.Data))
	}

	if retrieved.Metadata["project"] != "test" {
		t.Errorf("Expected metadata project='test', got '%s'", retrieved.Metadata["project"])
	}

	// Test Delete
	err = storage.Delete("test-key")
	if err != nil {
		t.Fatalf("Failed to delete cache entry: %v", err)
	}

	// Verify deletion
	_, err = storage.Get("test-key")
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestFileSystemStorage_ExpiredEntry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage with very short TTL
	storage := NewFileSystemStorage(tempDir, time.Millisecond*100)

	entry := &CacheEntry{
		Key:       "expired-key",
		Data:      []byte("expired data"),
		Timestamp: time.Now().Add(-time.Hour), // Already expired
		TTL:       time.Millisecond * 100,
		Metadata:  map[string]string{},
	}

	err = storage.Put("expired-key", entry)
	if err != nil {
		t.Fatalf("Failed to put expired entry: %v", err)
	}

	// Should get expired error
	_, err = storage.Get("expired-key")
	if err == nil {
		t.Error("Expected error for expired entry, got nil")
	} else if err.Error() != "cache entry expired" {
		t.Errorf("Expected 'cache entry expired' error, got: %v", err)
	}
}

func TestFileSystemStorage_List(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewFileSystemStorage(tempDir, time.Hour)

	// Add multiple entries
	for i := 0; i < 3; i++ {
		entry := &CacheEntry{
			Key:       fmt.Sprintf("key-%d", i),
			Data:      []byte(fmt.Sprintf("data-%d", i)),
			Timestamp: time.Now(),
			TTL:       time.Hour,
			Metadata:  map[string]string{},
		}
		err = storage.Put(fmt.Sprintf("key-%d", i), entry)
		if err != nil {
			t.Fatalf("Failed to put entry %d: %v", i, err)
		}
	}

	keys, err := storage.List()
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for i := 0; i < 3; i++ {
		expectedKey := fmt.Sprintf("key-%d", i)
		if !keyMap[expectedKey] {
			t.Errorf("Expected key %s not found in list", expectedKey)
		}
	}
}

func TestFileSystemStorage_Size(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewFileSystemStorage(tempDir, time.Hour)

	// Add entries with different sizes
	entries := []struct {
		key  string
		data string
	}{
		{"key1", "data1"},
		{"key2", "longer data for key2"},
		{"key3", "even longer data for key3 with more content"},
	}

	for _, e := range entries {
		entry := &CacheEntry{
			Key:       e.key,
			Data:      []byte(e.data),
			Timestamp: time.Now(),
			TTL:       time.Hour,
			Metadata:  map[string]string{},
		}
		err = storage.Put(e.key, entry)
		if err != nil {
			t.Fatalf("Failed to put entry %s: %v", e.key, err)
		}
	}

	size, err := storage.Size()
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}

	if size <= 0 {
		t.Errorf("Expected positive size, got %d", size)
	}
}

func TestCacheServer_GetPutDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100 * 1024 * 1024,
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	// Test Put
	err = server.Put("test-key", []byte("test data"), map[string]string{"project": "test"})
	if err != nil {
		t.Fatalf("Failed to put cache entry: %v", err)
	}

	// Test Get
	entry, err := server.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get cache entry: %v", err)
	}

	if string(entry.Data) != "test data" {
		t.Errorf("Expected data 'test data', got '%s'", string(entry.Data))
	}

	if entry.Metadata["project"] != "test" {
		t.Errorf("Expected metadata project='test', got '%s'", entry.Metadata["project"])
	}

	// Test Delete
	err = server.Delete("test-key")
	if err != nil {
		t.Fatalf("Failed to delete cache entry: %v", err)
	}

	// Verify deletion
	_, err = server.Get("test-key")
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestCacheServer_HandleCacheRequests(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100 * 1024 * 1024,
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	// Test PUT request
	putData := map[string]interface{}{
		"data": []byte("test data"),
		"metadata": map[string]string{
			"project": "test",
		},
	}
	putBody, _ := json.Marshal(putData)

	req := httptest.NewRequest(http.MethodPut, "/api/cache/test-key", bytes.NewReader(putBody))
	w := httptest.NewRecorder()
	server.handleCacheRequests(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for PUT, got %d", w.Code)
	}

	// Test GET request
	req = httptest.NewRequest(http.MethodGet, "/api/cache/test-key", nil)
	w = httptest.NewRecorder()
	server.handleCacheRequests(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for GET, got %d", w.Code)
	}

	var retrievedEntry CacheEntry
	err = json.NewDecoder(w.Body).Decode(&retrievedEntry)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if string(retrievedEntry.Data) != "test data" {
		t.Errorf("Expected data 'test data', got '%s'", string(retrievedEntry.Data))
	}

	// Test DELETE request
	req = httptest.NewRequest(http.MethodDelete, "/api/cache/test-key", nil)
	w = httptest.NewRecorder()
	server.handleCacheRequests(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for DELETE, got %d", w.Code)
	}

	// Verify deletion with GET
	req = httptest.NewRequest(http.MethodGet, "/api/cache/test-key", nil)
	w = httptest.NewRecorder()
	server.handleCacheRequests(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 after deletion, got %d", w.Code)
	}
}

func TestCacheServer_HandleMetrics(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100 * 1024 * 1024,
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	// Add some data to generate metrics
	server.Put("key1", []byte("data1"), map[string]string{})
	server.Get("key1")
	server.Get("non-existent") // Should increment misses

	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	w := httptest.NewRecorder()
	server.HandleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for metrics, got %d", w.Code)
	}

	var metrics CacheMetrics
	err = json.NewDecoder(w.Body).Decode(&metrics)
	if err != nil {
		t.Fatalf("Failed to decode metrics: %v", err)
	}

	if metrics.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", metrics.Hits)
	}

	if metrics.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", metrics.Misses)
	}

	if metrics.Operations["get"] != 2 {
		t.Errorf("Expected 2 get operations, got %d", metrics.Operations["get"])
	}

	if metrics.Operations["put"] != 1 {
		t.Errorf("Expected 1 put operation, got %d", metrics.Operations["put"])
	}
}

func TestCacheServer_HandleHealth(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100 * 1024 * 1024,
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	server.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for health, got %d", w.Code)
	}

	var health map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&health)
	if err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if status, ok := health["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

func TestCacheServer_CleanupWithML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100, // Very small cache size to trigger cleanup
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	// Add entries that exceed cache size
	for i := 0; i < 10; i++ {
		data := make([]byte, 50) // 50 bytes each
		metadata := map[string]string{
			"project_path": fmt.Sprintf("/projects/test-%d", i),
			"task_name":    fmt.Sprintf("build-%d", i),
		}
		err = server.Put(fmt.Sprintf("key-%d", i), data, metadata)
		if err != nil {
			t.Fatalf("Failed to put entry %d: %v", i, err)
		}
	}

	// Run cleanup
	err = server.CleanupWithML()
	if err != nil {
		t.Fatalf("CleanupWithML failed: %v", err)
	}

	// Check that some entries were evicted
	entries, err := server.Storage.List()
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if len(entries) >= 10 {
		t.Error("Expected some entries to be evicted, but all remain")
	}

	if server.Metrics.Evictions == 0 {
		t.Error("Expected evictions to be recorded")
	}
}

func TestCacheEntry_JSONSerialization(t *testing.T) {
	original := &CacheEntry{
		Key:       "test-key",
		Data:      []byte("test data"),
		Timestamp: time.Now().Truncate(time.Millisecond), // Truncate to avoid nanosecond precision issues
		TTL:       time.Hour,
		Metadata: map[string]string{
			"project": "test",
			"task":    "build",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal CacheEntry: %v", err)
	}

	// Unmarshal from JSON
	var decoded CacheEntry
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CacheEntry: %v", err)
	}

	// Compare
	if decoded.Key != original.Key {
		t.Errorf("Key mismatch: expected %s, got %s", original.Key, decoded.Key)
	}

	if string(decoded.Data) != string(original.Data) {
		t.Errorf("Data mismatch: expected %s, got %s", string(original.Data), string(decoded.Data))
	}

	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", original.Timestamp, decoded.Timestamp)
	}

	if decoded.TTL != original.TTL {
		t.Errorf("TTL mismatch: expected %v, got %v", original.TTL, decoded.TTL)
	}

	if len(decoded.Metadata) != len(original.Metadata) {
		t.Errorf("Metadata length mismatch: expected %d, got %d", len(original.Metadata), len(decoded.Metadata))
	}

	for k, v := range original.Metadata {
		if decoded.Metadata[k] != v {
			t.Errorf("Metadata[%s] mismatch: expected %s, got %s", k, v, decoded.Metadata[k])
		}
	}
}

func TestFileSystemStorage_Cleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage with short TTL
	storage := NewFileSystemStorage(tempDir, time.Millisecond*100)

	// Add expired entry
	expiredEntry := &CacheEntry{
		Key:       "expired",
		Data:      []byte("expired data"),
		Timestamp: time.Now().Add(-time.Hour),
		TTL:       time.Millisecond * 100,
		Metadata:  map[string]string{},
	}
	err = storage.Put("expired", expiredEntry)
	if err != nil {
		t.Fatalf("Failed to put expired entry: %v", err)
	}

	// Add valid entry
	validEntry := &CacheEntry{
		Key:       "valid",
		Data:      []byte("valid data"),
		Timestamp: time.Now(),
		TTL:       time.Hour,
		Metadata:  map[string]string{},
	}
	err = storage.Put("valid", validEntry)
	if err != nil {
		t.Fatalf("Failed to put valid entry: %v", err)
	}

	// Run cleanup
	err = storage.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Check that only valid entry remains
	keys, err := storage.List()
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 key after cleanup, got %d", len(keys))
	}

	if len(keys) > 0 && keys[0] != "valid" {
		t.Errorf("Expected 'valid' key, got %s", keys[0])
	}
}

func TestCacheServer_InvalidHTTPMethods(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100 * 1024 * 1024,
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	// Test POST method (not allowed)
	req := httptest.NewRequest(http.MethodPost, "/api/cache/test-key", nil)
	w := httptest.NewRecorder()
	server.handleCacheRequests(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for POST, got %d", w.Code)
	}

	// Test PATCH method (not allowed)
	req = httptest.NewRequest(http.MethodPatch, "/api/cache/test-key", nil)
	w = httptest.NewRecorder()
	server.handleCacheRequests(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for PATCH, got %d", w.Code)
	}
}

func TestCacheServer_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := types.CacheConfig{
		Port:            8080,
		StorageType:     "filesystem",
		StorageDir:      tempDir,
		MaxCacheSize:    100 * 1024 * 1024,
		TTL:             3600,
		CleanupInterval: time.Hour,
	}

	server := NewCacheServer(config)

	// Channel to coordinate goroutines
	done := make(chan bool)
	errors := make(chan error, 10)

	// Start multiple concurrent operations
	for i := 0; i < 5; i++ {
		go func(id int) {
			key := fmt.Sprintf("key-%d", id)
			data := []byte(fmt.Sprintf("data-%d", id))
			metadata := map[string]string{"id": fmt.Sprintf("%d", id)}

			// Put
			if err := server.Put(key, data, metadata); err != nil {
				errors <- fmt.Errorf("goroutine %d put failed: %v", id, err)
				return
			}

			// Get
			entry, err := server.Get(key)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d get failed: %v", id, err)
				return
			}

			if string(entry.Data) != string(data) {
				errors <- fmt.Errorf("goroutine %d data mismatch", id)
				return
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Fatal(err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify all entries are present
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key-%d", i)
		entry, err := server.Get(key)
		if err != nil {
			t.Errorf("Failed to get entry %s after concurrent operations: %v", key, err)
		}

		expectedData := fmt.Sprintf("data-%d", i)
		if string(entry.Data) != expectedData {
			t.Errorf("Data mismatch for %s: expected %s, got %s", key, expectedData, string(entry.Data))
		}
	}
}
