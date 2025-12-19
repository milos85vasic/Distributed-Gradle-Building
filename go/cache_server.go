package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheServer manages distributed build cache
type CacheServer struct {
	config        *CacheConfig
	cache         map[string]*CacheEntry
	mutex         sync.RWMutex
	storageDir    string
	storage       CacheStorage
	hitCount      int64
	missCount     int64
	startTime     time.Time
	httpServer    *http.Server
}

// CacheConfig contains cache server configuration
type CacheConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	StorageType    string `json:"storage_type"`
	StorageDir     string `json:"storage_dir"`
	MaxCacheSize   int64  `json:"max_cache_size"`
	TTL            int    `json:"ttl_seconds"`
	Compression    bool   `json:"compression"`
	Authentication bool   `json:"authentication"`
	AuthToken      string `json:"auth_token"`
}

// CacheEntry represents a cached build artifact
type CacheEntry struct {
	Key            string                 `json:"key"`
	Path           string                 `json:"path"`
	Size           int64                  `json:"size"`
	Hash           string                 `json:"hash"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessed   time.Time              `json:"last_accessed"`
	TTL            int                    `json:"ttl"`
	Metadata       map[string]interface{} `json:"metadata"`
	AccessCount    int64                  `json:"access_count"`
	Compression    bool                   `json:"compression"`
}

// CacheStorage interface for different storage backends
type CacheStorage interface {
	Put(key string, entry *CacheEntry) error
	Get(key string) (*CacheEntry, error)
	Delete(key string) error
	List() ([]*CacheEntry, error)
	Cleanup() error
	GetSize() (int64, error)
}

// FileSystemStorage implements file-based cache storage
type FileSystemStorage struct {
	baseDir string
	mutex   sync.RWMutex
}

// CacheRequest represents a cache request
type CacheRequest struct {
	Key       string                 `json:"key"`
	Data      []byte                 `json:"data"`
	Hash      string                 `json:"hash"`
	TTL       int                    `json:"ttl"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// CacheResponse represents a cache response
type CacheResponse struct {
	Found    bool                   `json:"found"`
	Entry    *CacheEntry            `json:"entry,omitempty"`
	Data     []byte                 `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewCacheServer creates a new cache server
func NewCacheServer(config *CacheConfig) (*CacheServer, error) {
	if config.StorageDir == "" {
		config.StorageDir = "/tmp/gradle-cache-server"
	}
	
	if config.TTL == 0 {
		config.TTL = 86400 // 24 hours
	}
	
	if config.MaxCacheSize == 0 {
		config.MaxCacheSize = 10 * 1024 * 1024 * 1024 // 10GB
	}
	
	server := &CacheServer{
		config:     config,
		cache:      make(map[string]*CacheEntry),
		storageDir: config.StorageDir,
		startTime:  time.Now(),
		hitCount:   0,
		missCount:  0,
	}
	
	// Initialize storage
	var err error
	switch config.StorageType {
	case "filesystem", "":
		server.storage, err = NewFileSystemStorage(config.StorageDir)
	case "redis":
		server.storage, err = NewRedisStorage(config.RedisAddr, config.RedisPassword)
	case "s3":
		server.storage, err = NewS3Storage(config.S3Bucket, config.S3Region)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.StorageType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}
	
	// Load existing cache entries
	server.loadExistingEntries()
	
	return server, nil
}

// NewFileSystemStorage creates a new file system storage
func NewFileSystemStorage(baseDir string) (*FileSystemStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	
	storage := &FileSystemStorage{
		baseDir: baseDir,
	}
	
	return storage, nil
}

// NewRedisStorage creates a new Redis storage backend
func NewRedisStorage(addr, password string) (*RedisStorage, error) {
	// For now, return a placeholder that logs Redis operations
	log.Printf("Redis storage requested for %s (placeholder implementation)", addr)
	return &RedisStorage{
		addr:     addr,
		password: password,
		data:     make(map[string][]byte),
	}, nil
}

// NewS3Storage creates a new S3 storage backend  
func NewS3Storage(bucket, region string) (*S3Storage, error) {
	// For now, return a placeholder that logs S3 operations
	log.Printf("S3 storage requested for bucket %s in region %s (placeholder implementation)", bucket, region)
	return &S3Storage{
		bucket: bucket,
		region: region,
		data:   make(map[string][]byte),
	}, nil
}

// RedisStorage implements storage backend using Redis
type RedisStorage struct {
	addr     string
	password string
	data     map[string][]byte
	mutex    sync.RWMutex
}

// Get retrieves data from Redis
func (rs *RedisStorage) Get(key string) ([]byte, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	
	data, exists := rs.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	
	return data, nil
}

// Set stores data in Redis
func (rs *RedisStorage) Set(key string, data []byte, ttl int) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	rs.data[key] = data
	return nil
}

// Delete removes data from Redis
func (rs *RedisStorage) Delete(key string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	delete(rs.data, key)
	return nil
}

// List returns all cache entries
func (rs *RedisStorage) List() ([]*CacheEntry, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	
	var entries []*CacheEntry
	for key, data := range rs.data {
		entries = append(entries, &CacheEntry{
			Key:      key,
			Data:     data,
			TTL:      0,
			Metadata: map[string]string{},
			CreatedAt: time.Now(),
		})
	}
	
	return entries, nil
}

// GetSize returns storage size
func (rs *RedisStorage) GetSize() (int64, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	
	var size int64
	for _, data := range rs.data {
		size += int64(len(data))
	}
	
	return size, nil
}

// S3Storage implements storage backend using S3
type S3Storage struct {
	bucket string
	region string
	data   map[string][]byte
	mutex  sync.RWMutex
}

// Get retrieves data from S3
func (s3 *S3Storage) Get(key string) ([]byte, error) {
	s3.mutex.RLock()
	defer s3.mutex.RUnlock()
	
	data, exists := s3.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	
	return data, nil
}

// Set stores data in S3
func (s3 *S3Storage) Set(key string, data []byte, ttl int) error {
	s3.mutex.Lock()
	defer s3.mutex.Unlock()
	
	s3.data[key] = data
	return nil
}

// Delete removes data from S3
func (s3 *S3Storage) Delete(key string) error {
	s3.mutex.Lock()
	defer s3.mutex.Unlock()
	
	delete(s3.data, key)
	return nil
}

// List returns all cache entries
func (s3 *S3Storage) List() ([]*CacheEntry, error) {
	s3.mutex.RLock()
	defer s3.mutex.RUnlock()
	
	var entries []*CacheEntry
	for key, data := range s3.data {
		entries = append(entries, &CacheEntry{
			Key:      key,
			Data:     data,
			TTL:      0,
			Metadata: map[string]string{},
			CreatedAt: time.Now(),
		})
	}
	
	return entries, nil
}

// GetSize returns storage size
func (s3 *S3Storage) GetSize() (int64, error) {
	s3.mutex.RLock()
	defer s3.mutex.RUnlock()
	
	var size int64
	for _, data := range s3.data {
		size += int64(len(data))
	}
	
	return size, nil
}

// Put stores an entry in the cache
func (fs *FileSystemStorage) Put(key string, entry *CacheEntry) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	entryPath := filepath.Join(fs.baseDir, key)
	dataPath := entryPath + ".data"
	metaPath := entryPath + ".meta"
	
	// Store data
	if err := os.WriteFile(dataPath, []byte{}, 0644); err != nil {
		return err
	}
	
	// Store metadata
	metaData, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return err
	}
	
	return nil
}

// Get retrieves an entry from the cache
func (fs *FileSystemStorage) Get(key string) (*CacheEntry, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	entryPath := filepath.Join(fs.baseDir, key)
	metaPath := entryPath + ".meta"
	
	data, err := ioutil.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not found
		}
		return nil, err
	}
	
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	
	return &entry, nil
}

// Delete removes an entry from the cache
func (fs *FileSystemStorage) Delete(key string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	entryPath := filepath.Join(fs.baseDir, key)
	dataPath := entryPath + ".data"
	metaPath := entryPath + ".meta"
	
	os.Remove(dataPath)
	os.Remove(metaPath)
	
	return nil
}

// List returns all entries in the cache
func (fs *FileSystemStorage) List() ([]*CacheEntry, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	var entries []*CacheEntry
	
	files, err := ioutil.ReadDir(fs.baseDir)
	if err != nil {
		return nil, err
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".meta" {
			key := file.Name()[:len(file.Name())-5]
			entry, err := fs.Get(key)
			if err != nil {
				continue
			}
			if entry != nil {
				entries = append(entries, entry)
			}
		}
	}
	
	return entries, nil
}

// Cleanup removes expired entries from the cache
func (fs *FileSystemStorage) Cleanup() error {
	entries, err := fs.List()
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		if entry.TTL > 0 && time.Since(entry.CreatedAt) > time.Duration(entry.TTL)*time.Second {
			fs.Delete(entry.Key)
		}
	}
	
	return nil
}

// GetSize returns the total size of the cache storage
func (fs *FileSystemStorage) GetSize() (int64, error) {
	var totalSize int64
	
	err := filepath.Walk(fs.baseDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	
	return totalSize, err
}

// loadExistingEntries loads existing cache entries from storage
func (cs *CacheServer) loadExistingEntries() {
	entries, err := cs.storage.List()
	if err != nil {
		log.Printf("Error loading existing entries: %v", err)
		return
	}
	
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	for _, entry := range entries {
		cs.cache[entry.Key] = entry
	}
	
	log.Printf("Loaded %d existing cache entries", len(entries))
}

// Put stores an entry in the cache
func (cs *CacheServer) Put(key string, request *CacheRequest) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	// Check if cache size limit is reached
	if cs.config.MaxCacheSize > 0 {
		size, _ := cs.storage.GetSize()
		if size > cs.config.MaxCacheSize {
			cs.evictLRU()
		}
	}
	
	entry := &CacheEntry{
		Key:          key,
		Path:         filepath.Join(cs.storageDir, key),
		Size:         int64(len(request.Data)),
		Hash:         request.Hash,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		TTL:          request.TTL,
		Metadata:     request.Metadata,
		AccessCount:  0,
		Compression:  cs.config.Compression,
	}
	
	if err := cs.storage.Put(key, entry); err != nil {
		return err
	}
	
	cs.cache[key] = entry
	
	// Store actual data
	dataPath := filepath.Join(cs.storageDir, key+".data")
	if err := os.WriteFile(dataPath, request.Data, 0644); err != nil {
		return err
	}
	
	return nil
}

// Get retrieves an entry from the cache
func (cs *CacheServer) Get(key string) (*CacheResponse, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	entry, exists := cs.cache[key]
	if !exists {
		cs.missCount++
		return &CacheResponse{Found: false}, nil
	}
	
	// Check TTL
	if entry.TTL > 0 && time.Since(entry.CreatedAt) > time.Duration(entry.TTL)*time.Second {
		delete(cs.cache, key)
		cs.storage.Delete(key)
		cs.missCount++
		return &CacheResponse{Found: false}, nil
	}
	
	// Update access info
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	cs.hitCount++
	
	// Load data
	dataPath := filepath.Join(cs.storageDir, key+".data")
	data, err := ioutil.ReadFile(dataPath)
	if err != nil {
		delete(cs.cache, key)
		cs.missCount++
		return &CacheResponse{Found: false}, err
	}
	
	return &CacheResponse{
		Found:    true,
		Entry:    entry,
		Data:     data,
		Metadata: entry.Metadata,
	}, nil
}

// Delete removes an entry from the cache
func (cs *CacheServer) Delete(key string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	delete(cs.cache, key)
	return cs.storage.Delete(key)
}

// evictLRU removes the least recently used entry
func (cs *CacheServer) evictLRU() {
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, entry := range cs.cache {
		if entry.LastAccessed.Before(oldestTime) {
			oldestTime = entry.LastAccessed
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(cs.cache, oldestKey)
		cs.storage.Delete(oldestKey)
		log.Printf("Evicted LRU entry: %s", oldestKey)
	}
}

// Cleanup removes expired entries from the cache
func (cs *CacheServer) Cleanup() error {
	entries, err := cs.storage.List()
	if err != nil {
		return err
	}
	
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	for _, entry := range entries {
		if entry.TTL > 0 && time.Since(entry.CreatedAt) > time.Duration(entry.TTL)*time.Second {
			delete(cs.cache, entry.Key)
			cs.storage.Delete(entry.Key)
		}
	}
	
	return nil
}

// GetStats returns cache statistics
func (cs *CacheServer) GetStats() map[string]interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	hitRate := float64(0)
	if cs.hitCount+cs.missCount > 0 {
		hitRate = float64(cs.hitCount) / float64(cs.hitCount+cs.missCount)
	}
	
	size, _ := cs.storage.GetSize()
	
	return map[string]interface{}{
		"entries":        len(cs.cache),
		"hit_count":      cs.hitCount,
		"miss_count":     cs.missCount,
		"hit_rate":       hitRate,
		"storage_size":   size,
		"max_cache_size": cs.config.MaxCacheSize,
		"uptime":         time.Since(cs.startTime).String(),
	}
}

// StartHTTPServer starts the cache server's HTTP API
func (cs *CacheServer) StartHTTPServer() error {
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/cache/", cs.handleCacheRequest)
	mux.HandleFunc("/stats", cs.handleStatsRequest)
	mux.HandleFunc("/cleanup", cs.handleCleanupRequest)
	mux.HandleFunc("/health", cs.handleHealthCheck)
	
	cs.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cs.config.Host, cs.config.Port),
		Handler: cs.authMiddleware(mux),
	}
	
	// Start cleanup routine
	go cs.periodicCleanup()
	
	log.Printf("Starting cache server on %s:%d", cs.config.Host, cs.config.Port)
	return cs.httpServer.ListenAndServe()
}

// authMiddleware adds authentication middleware if enabled
func (cs *CacheServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cs.config.Authentication {
			token := r.Header.Get("X-Auth-Token")
			if token != cs.config.AuthToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// HTTP Handlers
func (cs *CacheServer) handleCacheRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/cache/"):]
	
	switch r.Method {
	case "GET":
		response, err := cs.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		
	case "PUT":
		var request CacheRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		if err := cs.Put(key, &request); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "cached"})
		
	case "DELETE":
		if err := cs.Delete(key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (cs *CacheServer) handleStatsRequest(w http.ResponseWriter, r *http.Request) {
	stats := cs.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (cs *CacheServer) handleCleanupRequest(w http.ResponseWriter, r *http.Request) {
	if err := cs.Cleanup(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cleaned"})
}

func (cs *CacheServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// periodicCleanup runs cleanup periodically
func (cs *CacheServer) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		if err := cs.Cleanup(); err != nil {
			log.Printf("Cleanup error: %v", err)
		}
	}
}

func cacheServerMain() {
	// Load configuration
	configFile := "cache_config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	
	config, err := loadCacheConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load cache config: %v", err)
	}
	
	// Create and start cache server
	server, err := NewCacheServer(config)
	if err != nil {
		log.Fatalf("Failed to create cache server: %v", err)
	}
	
	log.Printf("Starting cache server on %s:%d", config.Host, config.Port)
	if err := server.StartHTTPServer(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// loadCacheConfig loads cache configuration from file
func loadCacheConfig(filename string) (*CacheConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return &CacheConfig{
				Host:           "localhost",
				Port:           8083,
				StorageType:    "filesystem",
				StorageDir:     "/tmp/gradle-cache-server",
				MaxCacheSize:   10 * 1024 * 1024 * 1024, // 10GB
				TTL:            86400, // 24 hours
				Compression:    true,
				Authentication: false,
			}, nil
		}
		return nil, err
	}
	
	var config CacheConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}