package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// Test cache server functionality
func TestCacheServer_Creation(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	if cache == nil {
		t.Fatal("Failed to create cache server")
	}
	
	if cache.config.StorageType != config.StorageType {
		t.Errorf("Expected storage type %s, got %s", config.StorageType, cache.config.StorageType)
	}
	
	if cache.config.MaxSize != config.MaxSize {
		t.Errorf("Expected max size %d, got %d", config.MaxSize, cache.config.MaxSize)
	}
}

func TestCacheServer_StoreAndRetrieve(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Test data
	key := "test-cache-key"
	data := CacheEntry{
		Key:        key,
		Data:       []byte("test data"),
		Metadata:   map[string]string{"type": "gradle-cache"},
		CreatedAt:  time.Now(),
		AccessCount: 0,
		Size:       9,
	}
	
	// Store data
	err := cache.Store(key, data)
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Retrieve data
	retrieved, err := cache.Retrieve(key)
	if err != nil {
		t.Fatalf("Failed to retrieve cache entry: %v", err)
	}
	
	if retrieved.Key != key {
		t.Errorf("Expected key %s, got %s", key, retrieved.Key)
	}
	
	if string(retrieved.Data) != "test data" {
		t.Errorf("Expected data 'test data', got '%s'", string(retrieved.Data))
	}
	
	if retrieved.AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", retrieved.AccessCount)
	}
}

func TestCacheServer_Delete(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Store test data
	key := "test-delete-key"
	data := CacheEntry{
		Key:  key,
		Data: []byte("test data"),
	}
	
	err := cache.Store(key, data)
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Delete data
	err = cache.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete cache entry: %v", err)
	}
	
	// Verify deletion
	_, err = cache.Retrieve(key)
	if err == nil {
		t.Error("Expected error when retrieving deleted entry")
	}
}

func TestCacheServer_CleanupExpired(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          1, // 1 second TTL
		CleanupInterval: 1,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Store test data
	key := "expired-key"
	data := CacheEntry{
		Key:       key,
		Data:      []byte("test data"),
		CreatedAt: time.Now().Add(-2 * time.Second), // Created 2 seconds ago
	}
	
	err := cache.Store(key, data)
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Wait for expiration
	time.Sleep(2 * time.Second)
	
	// Run cleanup
	cache.cleanupExpired()
	
	// Verify expired entry was removed
	_, err = cache.Retrieve(key)
	if err == nil {
		t.Error("Expected error when retrieving expired entry")
	}
}

func TestCacheServer_SizeLimit(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      10, // 10 bytes limit
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Store data within limit
	key1 := "key1"
	data1 := CacheEntry{
		Key:  key1,
		Data: []byte("small"), // 5 bytes
		Size: 5,
	}
	
	err := cache.Store(key1, data1)
	if err != nil {
		t.Fatalf("Failed to store small cache entry: %v", err)
	}
	
	// Store data that exceeds limit
	key2 := "key2"
	data2 := CacheEntry{
		Key:  key2,
		Data: []byte("this is too large for the cache"), // >10 bytes
		Size: 33,
	}
	
	err = cache.Store(key2, data2)
	if err == nil {
		t.Error("Expected error when storing oversized entry")
	}
}

func TestCacheServer_HTTPHandlers(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Test health check
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	cache.handleHealth(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "ok" {
		t.Errorf("Expected ok status, got %s", response["status"])
	}
	
	// Test cache GET - missing key
	req, _ = http.NewRequest("GET", "/cache/nonexistent", nil)
	w = httptest.NewRecorder()
	
	cache.handleGetCache(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	
	// Test cache POST
	entryData := map[string]interface{}{
		"key":  "test-key",
		"data": "dGVzdCBkYXRh", // base64 encoded "test data"
		"metadata": map[string]string{"type": "test"},
	}
	
	jsonData, _ := json.Marshal(entryData)
	req, _ = http.NewRequest("POST", "/cache", httptest.NewRecorder())
	req.Body = httptest.NewRecorder().Body
	req.Body = httptest.NewRecorder().Body
	
	w = httptest.NewRecorder()
	cache.handleStoreCache(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected bad request for empty body, got %d", w.Code)
	}
}

func TestCacheServer_Statistics(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Store some test data
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		data := CacheEntry{
			Key:  key,
			Data: []byte(fmt.Sprintf("test data %d", i)),
			Size: len(fmt.Sprintf("test data %d", i)),
		}
		
		err := cache.Store(key, data)
		if err != nil {
			t.Fatalf("Failed to store cache entry %d: %v", i, err)
		}
	}
	
	// Get statistics
	stats := cache.GetStatistics()
	
	if stats.TotalEntries != 5 {
		t.Errorf("Expected 5 total entries, got %d", stats.TotalEntries)
	}
	
	if stats.TotalSize <= 0 {
		t.Error("Expected positive total size")
	}
	
	if stats.HitRate != 0.0 {
		t.Errorf("Expected hit rate 0.0, got %f", stats.HitRate)
	}
}

func TestCacheConfig_Load(t *testing.T) {
	// Create test config file
	testConfig := CacheConfig{
		StorageType:     "disk",
		MaxSize:         1000,
		TTL:             7200,
		CleanupInterval:  600,
		Port:            8084,
		StoragePath:     "/tmp/cache",
		CompressionEnabled: true,
	}
	
	configJSON, _ := json.Marshal(testConfig)
	testConfigPath := "/tmp/test_cache_config.json"
	
	err := ioutil.WriteFile(testConfigPath, configJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove(testConfigPath)
	
	// Load config
	loadedConfig, err := loadCacheConfig(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load cache config: %v", err)
	}
	
	if loadedConfig.StorageType != testConfig.StorageType {
		t.Errorf("Expected storage type %s, got %s", testConfig.StorageType, loadedConfig.StorageType)
	}
	
	if loadedConfig.MaxSize != testConfig.MaxSize {
		t.Errorf("Expected max size %d, got %d", testConfig.MaxSize, loadedConfig.MaxSize)
	}
	
	if loadedConfig.Port != testConfig.Port {
		t.Errorf("Expected port %d, got %d", testConfig.Port, loadedConfig.Port)
	}
}

func TestCacheServer_BackupAndRestore(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      100,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
		StoragePath:  "/tmp/backup-test",
	}
	
	cache := NewCacheServer(config)
	
	// Store test data
	key := "backup-test-key"
	data := CacheEntry{
		Key:      key,
		Data:     []byte("backup test data"),
		Metadata: map[string]string{"backup": "true"},
	}
	
	err := cache.Store(key, data)
	if err != nil {
		t.Fatalf("Failed to store cache entry: %v", err)
	}
	
	// Create backup
	backupPath := "/tmp/cache-backup.json"
	err = cache.CreateBackup(backupPath)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	defer os.Remove(backupPath)
	
	// Create new cache and restore
	newCache := NewCacheServer(config)
	err = newCache.RestoreFromBackup(backupPath)
	if err != nil {
		t.Fatalf("Failed to restore from backup: %v", err)
	}
	
	// Verify restored data
	retrieved, err := newCache.Retrieve(key)
	if err != nil {
		t.Fatalf("Failed to retrieve restored entry: %v", err)
	}
	
	if string(retrieved.Data) != "backup test data" {
		t.Errorf("Expected 'backup test data', got '%s'", string(retrieved.Data))
	}
}

func TestCacheServer_ConcurrentAccess(t *testing.T) {
	config := &CacheConfig{
		StorageType:  "memory",
		MaxSize:      1000,
		TTL:          3600,
		CleanupInterval: 300,
		Port:         8084,
	}
	
	cache := NewCacheServer(config)
	
	// Test concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			key := fmt.Sprintf("concurrent-key-%d", id)
			data := CacheEntry{
				Key:  key,
				Data: []byte(fmt.Sprintf("concurrent data %d", id)),
			}
			
			err := cache.Store(key, data)
			if err != nil {
				t.Errorf("Failed to store concurrent entry %d: %v", id, err)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify all entries were stored
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("concurrent-key-%d", i)
		_, err := cache.Retrieve(key)
		if err != nil {
			t.Errorf("Failed to retrieve concurrent entry %d: %v", i, err)
		}
	}
}

func TestCacheServer_EvictionPolicy(t *testing.T) {
	config := &CacheConfig{
		StorageType:   "memory",
		MaxSize:       20, // 20 bytes limit
		TTL:           3600,
		CleanupInterval: 300,
		Port:          8084,
		EvictionPolicy: "LRU", // Least Recently Used
	}
	
	cache := NewCacheServer(config)
	
	// Fill cache to capacity
	for i := 0; i < 4; i++ {
		key := fmt.Sprintf("evict-key-%d", i)
		data := CacheEntry{
			Key:  key,
			Data: []byte(fmt.Sprintf("data-%d", i)),
			Size: 6, // Each entry is 6 bytes
		}
		
		err := cache.Store(key, data)
		if err != nil {
			t.Fatalf("Failed to store entry %d: %v", i, err)
		}
	}
	
	// Access first few entries to make them recently used
	for i := 0; i < 2; i++ {
		key := fmt.Sprintf("evict-key-%d", i)
		_, err := cache.Retrieve(key)
		if err != nil {
			t.Fatalf("Failed to retrieve entry %d: %v", i, err)
		}
	}
	
	// Add one more entry that should trigger eviction
	key := "evict-key-4"
	data := CacheEntry{
		Key:  key,
		Data: []byte("data-4"),
		Size: 6,
	}
	
	err := cache.Store(key, data)
	if err != nil {
		t.Fatalf("Failed to store new entry: %v", err)
	}
	
	// Check that least recently used entries were evicted
	stats := cache.GetStatistics()
	if stats.TotalEntries > 4 {
		t.Errorf("Expected 4 or fewer entries after eviction, got %d", stats.TotalEntries)
	}
	
	// Verify most recently used entries are still there
	for i := 0; i < 2; i++ {
		key := fmt.Sprintf("evict-key-%d", i)
		_, err := cache.Retrieve(key)
		if err != nil {
			t.Errorf("Recently used entry %d was evicted", i)
		}
	}
}