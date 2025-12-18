package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestCacheServerStartup tests cache server startup and initialization
func TestCacheServerStartup(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server configuration
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  1 * time.Hour,
		TTL:             24 * time.Hour,
		Compression:     true,
		Encryption:       false,
	}
	
	// Create cache server instance
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Verify cache initialization
	if cache.Host != config.Host {
		t.Errorf("Expected host '%s', got '%s'", config.Host, cache.Host)
	}
	
	if cache.Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, cache.Port)
	}
	
	if cache.Status != "initializing" {
		t.Errorf("Expected status 'initializing', got '%s'", cache.Status)
	}
	
	if cache.MaxSize != int64(config.MaxSizeMB)*1024*1024 {
		t.Errorf("Expected max size %d, got %d", int64(config.MaxSizeMB)*1024*1024, cache.MaxSize)
	}
}

// TestCacheEntryStorage tests cache entry storage and retrieval
func TestCacheEntryStorage(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server instance
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  1 * time.Hour,
		TTL:             24 * time.Hour,
		Compression:     false,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Test data
	testKey := "test-key-123"
	testData := []byte("test cache data for verification")
	
	// Store cache entry
	err = cache.Store(testKey, testData, map[string]string{
		"build-id": "build-456",
		"task-type": "compileJava",
	})
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Retrieve cache entry
	retrievedData, metadata, err := cache.Retrieve(testKey)
	if err != nil {
		t.Errorf("Failed to retrieve cache entry: %v", err)
	}
	
	// Verify retrieved data
	if !bytes.Equal(retrievedData, testData) {
		t.Error("Retrieved data does not match stored data")
	}
	
	// Verify metadata
	if metadata["build-id"] != "build-456" {
		t.Errorf("Expected build-id 'build-456', got '%s'", metadata["build-id"])
	}
	
	if metadata["task-type"] != "compileJava" {
		t.Errorf("Expected task-type 'compileJava', got '%s'", metadata["task-type"])
	}
	
	// Verify timestamp
	if metadata["timestamp"] == "" {
		t.Error("Timestamp should be set in metadata")
	}
	
	// Verify size
	if metadata["size"] == "" {
		t.Error("Size should be set in metadata")
	}
}

// TestCacheEviction tests cache eviction when size limit is reached
func TestCacheEviction(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server with small size limit
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1, // 1MB limit
		CleanupInterval:  1 * time.Hour,
		TTL:             24 * time.Hour,
		Compression:     false,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Store multiple entries to exceed size limit
	largeData := make([]byte, 300*1024) // 300KB each
	
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("large-entry-%d", i)
		err := cache.Store(key, largeData, map[string]string{
			"index": fmt.Sprintf("%d", i),
		})
		if err != nil {
			t.Errorf("Failed to store entry %d: %v", i, err)
		}
	}
	
	// Verify that some entries were evicted (LRU eviction)
	stats := cache.GetStatistics()
	
	if stats.TotalEntries == 0 {
		t.Error("At least one entry should remain after eviction")
	}
	
	if stats.TotalSize > int64(config.MaxSizeMB)*1024*1024 {
		t.Errorf("Cache size %d should not exceed limit %d", stats.TotalSize, int64(config.MaxSizeMB)*1024*1024)
	}
	
	// Test LRU behavior - first entries should be evicted
	_, _, err = cache.Retrieve("large-entry-0")
	if err == nil {
		t.Error("First entry should be evicted due to LRU policy")
	}
	
	// Last entry should still be available
	_, _, err = cache.Retrieve("large-entry-4")
	if err != nil {
		t.Errorf("Last entry should still be available: %v", err)
	}
}

// TestCacheExpiration tests cache entry expiration based on TTL
func TestCacheExpiration(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server with short TTL
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  1 * time.Hour,
		TTL:             1 * time.Second, // Very short TTL
		Compression:     false,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Store cache entry
	testKey := "test-expiration"
	testData := []byte("test data for expiration")
	
	err = cache.Store(testKey, testData, map[string]string{
		"test": "expiration",
	})
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Verify entry exists immediately
	retrievedData, _, err := cache.Retrieve(testKey)
	if err != nil {
		t.Errorf("Cache entry should be available immediately: %v", err)
	}
	if !bytes.Equal(retrievedData, testData) {
		t.Error("Retrieved data does not match stored data")
	}
	
	// Wait for expiration
	time.Sleep(2 * time.Second)
	
	// Verify entry expired
	_, _, err = cache.Retrieve(testKey)
	if err == nil {
		t.Error("Cache entry should be expired after TTL")
	}
}

// TestCacheCompression tests cache entry compression
func TestCacheCompression(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server with compression enabled
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  1 * time.Hour,
		TTL:             24 * time.Hour,
		Compression:     true,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Create compressible test data (repeating pattern)
	compressibleData := make([]byte, 10*1024)
	for i := range compressibleData {
		compressibleData[i] = byte(i % 256)
	}
	
	testKey := "test-compression"
	
	// Store cache entry
	err = cache.Store(testKey, compressibleData, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Retrieve cache entry
	retrievedData, metadata, err := cache.Retrieve(testKey)
	if err != nil {
		t.Errorf("Failed to retrieve cache entry: %v", err)
	}
	
	// Verify retrieved data matches original
	if !bytes.Equal(retrievedData, compressibleData) {
		t.Error("Retrieved compressed data does not match original")
	}
	
	// Verify compression metadata
	if metadata["compressed"] != "true" {
		t.Errorf("Expected compression flag 'true', got '%s'", metadata["compressed"])
	}
	
	// Verify size savings
	originalSize := len(compressibleData)
	storedSize, _ := strconv.Atoi(metadata["stored-size"])
	
	if storedSize >= originalSize {
		t.Errorf("Compressed size %d should be smaller than original %d", storedSize, originalSize)
	}
}

// TestCacheAPI tests cache server HTTP API
func TestCacheAPI(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  1 * time.Hour,
		TTL:             24 * time.Hour,
		Compression:     true,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Create router
	router := mux.NewRouter()
	router.HandleFunc("/api/cache/{key}", cache.GetCacheHandler).Methods("GET")
	router.HandleFunc("/api/cache/{key}", cache.StoreCacheHandler).Methods("PUT")
	router.HandleFunc("/api/cache/{key}", cache.DeleteCacheHandler).Methods("DELETE")
	router.HandleFunc("/api/cache/statistics", cache.GetStatisticsHandler).Methods("GET")
	router.HandleFunc("/api/cache/cleanup", cache.CleanupHandler).Methods("POST")
	
	// Test store endpoint
	testKey := "api-test-key"
	testData := map[string]interface{}{
		"message": "test data",
		"number":  123,
		"array":   []string{"a", "b", "c"},
	}
	
	reqBody, _ := json.Marshal(testData)
	req, _ := http.NewRequest("PUT", "/api/cache/"+testKey, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d for store, got %d", http.StatusCreated, status)
	}
	
	// Test retrieve endpoint
	req, _ = http.NewRequest("GET", "/api/cache/"+testKey, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for retrieve, got %d", http.StatusOK, status)
	}
	
	var retrievedData map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &retrievedData)
	if err != nil {
		t.Errorf("Failed to unmarshal retrieved data: %v", err)
	}
	
	if retrievedData["message"] != testData["message"] {
		t.Errorf("Expected message '%s', got '%s'", testData["message"], retrievedData["message"])
	}
	
	// Test statistics endpoint
	req, _ = http.NewRequest("GET", "/api/cache/statistics", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for statistics, got %d", http.StatusOK, status)
	}
	
	var stats map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &stats)
	if err != nil {
		t.Errorf("Failed to unmarshal statistics: %v", err)
	}
	
	if stats["total_entries"] == nil {
		t.Error("Statistics should include total_entries")
	}
	
	if stats["total_size"] == nil {
		t.Error("Statistics should include total_size")
	}
	
	// Test delete endpoint
	req, _ = http.NewRequest("DELETE", "/api/cache/"+testKey, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for delete, got %d", http.StatusOK, status)
	}
	
	// Verify deletion
	req, _ = http.NewRequest("GET", "/api/cache/"+testKey, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %d for deleted item, got %d", http.StatusNotFound, status)
	}
}

// TestCacheBackends tests different cache backend implementations
func TestCacheBackends(t *testing.T) {
	testCases := []struct {
		backend string
		config  CacheConfig
	}{
		{
			backend: "filesystem",
			config: CacheConfig{
				Host:            "localhost",
				Port:            8081,
				DataDir:         "",
				Backend:         "filesystem",
				MaxSizeMB:       1024,
				CleanupInterval:  1 * time.Hour,
				TTL:             24 * time.Hour,
				Compression:     false,
				Encryption:       false,
			},
		},
		{
			backend: "memory",
			config: CacheConfig{
				Host:            "localhost",
				Port:            8081,
				DataDir:         "",
				Backend:         "memory",
				MaxSizeMB:       1024,
				CleanupInterval:  1 * time.Hour,
				TTL:             24 * time.Hour,
				Compression:     false,
				Encryption:       false,
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.backend, func(t *testing.T) {
			// Create temporary directory for filesystem backend
			if tc.backend == "filesystem" {
				tempDir, err := os.MkdirTemp("", "cache-fs-test")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tempDir)
				tc.config.DataDir = tempDir
			}
			
			// Create cache server
			cache, err := NewCacheServer(tc.config)
			if err != nil {
				t.Fatalf("Failed to create cache server with %s backend: %v", tc.backend, err)
			}
			
			// Test basic operations
			testKey := fmt.Sprintf("test-%s-backend", tc.backend)
			testData := []byte(fmt.Sprintf("test data for %s backend", tc.backend))
			
			// Store
			err = cache.Store(testKey, testData, map[string]string{
				"backend": tc.backend,
			})
			if err != nil {
				t.Errorf("Failed to store with %s backend: %v", tc.backend, err)
			}
			
			// Retrieve
			retrievedData, metadata, err := cache.Retrieve(testKey)
			if err != nil {
				t.Errorf("Failed to retrieve with %s backend: %v", tc.backend, err)
			}
			
			if !bytes.Equal(retrievedData, testData) {
				t.Error("Retrieved data does not match stored data")
			}
			
			if metadata["backend"] != tc.backend {
				t.Errorf("Expected backend '%s', got '%s'", tc.backend, metadata["backend"])
			}
			
			// Verify statistics
			stats := cache.GetStatistics()
			if stats.TotalEntries == 0 {
				t.Errorf("Statistics should show entries for %s backend", tc.backend)
			}
		})
	}
}

// TestCacheCleanup tests automatic cache cleanup
func TestCacheCleanup(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server with short cleanup interval
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  500 * time.Millisecond, // Short cleanup interval
		TTL:             1 * time.Second, // Short TTL
		Compression:     false,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Start cleanup goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go cache.StartCleanup(ctx)
	
	// Store entries that will expire
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("cleanup-test-%d", i)
		data := []byte(fmt.Sprintf("test data %d", i))
		
		err := cache.Store(key, data, map[string]string{
			"index": fmt.Sprintf("%d", i),
		})
		if err != nil {
			t.Errorf("Failed to store entry %d: %v", i, err)
		}
	}
	
	// Verify entries exist immediately
	stats := cache.GetStatistics()
	if stats.TotalEntries != 5 {
		t.Errorf("Expected 5 entries initially, got %d", stats.TotalEntries)
	}
	
	// Wait for cleanup to run
	time.Sleep(2 * time.Second)
	
	// Verify entries were cleaned up
	stats = cache.GetStatistics()
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", stats.TotalEntries)
	}
}

// TestCacheConcurrentAccess tests concurrent access to cache
func TestCacheConcurrentAccess(t *testing.T) {
	// Create temporary directory for cache
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create cache server
	config := CacheConfig{
		Host:            "localhost",
		Port:            8081,
		DataDir:         tempDir,
		Backend:         "filesystem",
		MaxSizeMB:       1024,
		CleanupInterval:  1 * time.Hour,
		TTL:             24 * time.Hour,
		Compression:     false,
		Encryption:       false,
	}
	
	cache, err := NewCacheServer(config)
	if err != nil {
		t.Fatalf("Failed to create cache server: %v", err)
	}
	
	// Test parameters
	numGoroutines := 10
	numOperations := 100
	
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)
	
	// Start concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("concurrent-test-%d-%d", goroutineID, j)
				data := []byte(fmt.Sprintf("test data %d-%d", goroutineID, j))
				
				// Store
				err := cache.Store(key, data, map[string]string{
					"goroutine": fmt.Sprintf("%d", goroutineID),
					"operation": fmt.Sprintf("%d", j),
				})
				if err != nil {
					errors <- fmt.Errorf("store error: %v", err)
					continue
				}
				
				// Retrieve
				retrievedData, _, err := cache.Retrieve(key)
				if err != nil {
					errors <- fmt.Errorf("retrieve error: %v", err)
					continue
				}
				
				if !bytes.Equal(retrievedData, data) {
					errors <- fmt.Errorf("data mismatch for key %s", key)
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
		errorCount++
	}
	
	if errorCount > 0 {
		t.Errorf("Expected no errors, got %d", errorCount)
	}
	
	// Verify final statistics
	stats := cache.GetStatistics()
	expectedEntries := numGoroutines * numOperations
	if stats.TotalEntries != expectedEntries {
		t.Errorf("Expected %d entries, got %d", expectedEntries, stats.TotalEntries)
	}
}

// TestCacheValidation tests cache configuration and input validation
func TestCacheValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      CacheConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: CacheConfig{
				Host:            "localhost",
				Port:            8081,
				DataDir:         "/tmp/cache",
				Backend:         "filesystem",
				MaxSizeMB:       1024,
				CleanupInterval:  1 * time.Hour,
				TTL:             24 * time.Hour,
				Compression:     true,
				Encryption:       false,
			},
			expectError: false,
		},
		{
			name: "Invalid backend",
			config: CacheConfig{
				Host:            "localhost",
				Port:            8081,
				DataDir:         "/tmp/cache",
				Backend:         "invalid",
				MaxSizeMB:       1024,
				CleanupInterval:  1 * time.Hour,
				TTL:             24 * time.Hour,
				Compression:     true,
				Encryption:       false,
			},
			expectError: true,
		},
		{
			name: "Zero max size",
			config: CacheConfig{
				Host:            "localhost",
				Port:            8081,
				DataDir:         "/tmp/cache",
				Backend:         "filesystem",
				MaxSizeMB:       0,
				CleanupInterval:  1 * time.Hour,
				TTL:             24 * time.Hour,
				Compression:     true,
				Encryption:       false,
			},
			expectError: true,
		},
		{
			name: "Negative TTL",
			config: CacheConfig{
				Host:            "localhost",
				Port:            8081,
				DataDir:         "/tmp/cache",
				Backend:         "filesystem",
				MaxSizeMB:       1024,
				CleanupInterval:  1 * time.Hour,
				TTL:             -1 * time.Second,
				Compression:     true,
				Encryption:       false,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache, err := NewCacheServer(tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but cache creation succeeded")
				}
				if cache != nil {
					t.Error("Expected nil cache on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if cache == nil {
					t.Error("Expected non-nil cache on success")
				}
			}
		})
	}
}