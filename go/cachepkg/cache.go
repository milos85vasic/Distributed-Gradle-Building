package cachepkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"distributed-gradle-building/types"
)

// CacheServer represents a distributed build cache server
type CacheServer struct {
	Config      types.CacheConfig
	Storage     CacheStorage
	Metrics     CacheMetrics
	Mutex       sync.RWMutex
	httpServer  *http.Server
	shutdown    chan struct{}
}

// CacheStorage interface for different cache storage backends
type CacheStorage interface {
	Get(key string) (*CacheEntry, error)
	Put(key string, entry *CacheEntry) error
	Delete(key string) error
	List() ([]string, error)
	Size() (int64, error)
	Cleanup() error
}

// CacheEntry represents a cached build artifact
type CacheEntry struct {
	Key       string            `json:"key"`
	Data      []byte            `json:"data"`
	Timestamp time.Time         `json:"timestamp"`
	TTL       time.Duration     `json:"ttl"`
	Metadata  map[string]string `json:"metadata"`
}

// CacheMetrics contains cache performance metrics
type CacheMetrics struct {
	Hits        int64            `json:"hits"`
	Misses      int64            `json:"misses"`
	Size        int64            `json:"size"`
	Entries     int              `json:"entries"`
	Evictions   int64            `json:"evictions"`
	LastCleanup time.Time        `json:"last_cleanup"`
	Operations  map[string]int64 `json:"operations"`
}

// FileSystemStorage implements file system based cache storage
type FileSystemStorage struct {
	BaseDir string
	TTL     time.Duration
}

// NewFileSystemStorage creates a new file system storage
func NewFileSystemStorage(baseDir string, ttl time.Duration) *FileSystemStorage {
	return &FileSystemStorage{
		BaseDir: baseDir,
		TTL:     ttl,
	}
}

// Get retrieves a cache entry from file system
func (fs *FileSystemStorage) Get(key string) (*CacheEntry, error) {
	filename := filepath.Join(fs.BaseDir, key+".cache")
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var entry CacheEntry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entry); err != nil {
		return nil, err
	}
	
	// Check TTL
	if time.Since(entry.Timestamp) > entry.TTL {
		os.Remove(filename)
		return nil, fmt.Errorf("cache entry expired")
	}
	
	return &entry, nil
}

// Put stores a cache entry to file system
func (fs *FileSystemStorage) Put(key string, entry *CacheEntry) error {
	if err := os.MkdirAll(fs.BaseDir, 0755); err != nil {
		return err
	}
	
	filename := filepath.Join(fs.BaseDir, key+".cache")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	return encoder.Encode(entry)
}

// Delete removes a cache entry from file system
func (fs *FileSystemStorage) Delete(key string) error {
	filename := filepath.Join(fs.BaseDir, key+".cache")
	return os.Remove(filename)
}

// List returns all cache keys
func (fs *FileSystemStorage) List() ([]string, error) {
	var keys []string
	
	err := filepath.Walk(fs.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && filepath.Ext(path) == ".cache" {
			key := filepath.Base(path)
			key = key[:len(key)-6] // Remove .cache extension
			keys = append(keys, key)
		}
		
		return nil
	})
	
	return keys, err
}

// Size returns the total size of cache storage
func (fs *FileSystemStorage) Size() (int64, error) {
	var totalSize int64
	
	err := filepath.Walk(fs.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && filepath.Ext(path) == ".cache" {
			totalSize += info.Size()
		}
		
		return nil
	})
	
	return totalSize, err
}

// Cleanup removes expired cache entries
func (fs *FileSystemStorage) Cleanup() error {
	entries, err := fs.List()
	if err != nil {
		return err
	}
	
	for _, key := range entries {
		entry, err := fs.Get(key)
		if err != nil {
			// Remove invalid entries
			fs.Delete(key)
			continue
		}
		
		if time.Since(entry.Timestamp) > fs.TTL {
			fs.Delete(key)
		}
	}
	
	return nil
}

// NewCacheServer creates a new cache server
func NewCacheServer(config types.CacheConfig) *CacheServer {
	var storage CacheStorage
	
	switch config.StorageType {
	case "filesystem":
		storage = NewFileSystemStorage(config.StorageDir, config.TTL)
	default:
		storage = NewFileSystemStorage(config.StorageDir, config.TTL)
	}
	
	return &CacheServer{
		Config:     config,
		Storage:    storage,
		Metrics:    CacheMetrics{Operations: make(map[string]int64)},
		shutdown:   make(chan struct{}),
	}
}

// Get retrieves a cache entry
func (cs *CacheServer) Get(key string) (*CacheEntry, error) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	
	entry, err := cs.Storage.Get(key)
	if err != nil {
		cs.Metrics.Misses++
		cs.Metrics.Operations["get"]++
		return nil, err
	}
	
	cs.Metrics.Hits++
	cs.Metrics.Operations["get"]++
	return entry, nil
}

// Put stores a cache entry
func (cs *CacheServer) Put(key string, data []byte, metadata map[string]string) error {
	cs.Mutex.Lock()
	
	entry := &CacheEntry{
		Key:       key,
		Data:      data,
		Timestamp: time.Now(),
		TTL:       cs.Config.TTL,
		Metadata:  metadata,
	}
	
	err := cs.Storage.Put(key, entry)
	if err != nil {
		cs.Mutex.Unlock()
		return err
	}
	
	cs.Metrics.Operations["put"]++
	cs.Mutex.Unlock()
	
	// Update metrics outside of lock to avoid deadlock
	cs.updateMetrics()
	return nil
}

// Delete removes a cache entry
func (cs *CacheServer) Delete(key string) error {
	cs.Mutex.Lock()
	
	err := cs.Storage.Delete(key)
	if err != nil {
		cs.Mutex.Unlock()
		return err
	}
	
	cs.Metrics.Operations["delete"]++
	cs.Mutex.Unlock()
	
	// Update metrics outside of lock to avoid deadlock
	cs.updateMetrics()
	return nil
}

// StartServer starts the HTTP server
func (cs *CacheServer) StartServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/cache/", cs.handleCacheRequests)
	mux.HandleFunc("/api/metrics", cs.handleMetrics)
	mux.HandleFunc("/api/health", cs.handleHealth)
	
	cs.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cs.Config.Port),
		Handler: mux,
	}
	
	go func() {
		log.Printf("Cache server listening on port %d", cs.Config.Port)
		if err := cs.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	
	// Start cleanup routine
	go cs.startCleanupRoutine()
	
	return nil
}

// Shutdown gracefully shuts down the cache server
func (cs *CacheServer) Shutdown() error {
	close(cs.shutdown)
	
	if cs.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cs.httpServer.Shutdown(ctx)
	}
	
	return nil
}

// handleCacheRequests handles cache-related HTTP requests
func (cs *CacheServer) handleCacheRequests(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/api/cache/"):]
	
	switch r.Method {
	case http.MethodGet:
		entry, err := cs.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entry)
		
	case http.MethodPut:
		var request struct {
			Data     []byte            `json:"data"`
			Metadata map[string]string `json:"metadata"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		if err := cs.Put(key, request.Data, request.Metadata); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		
	case http.MethodDelete:
		if err := cs.Delete(key); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMetrics handles metrics requests
func (cs *CacheServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	cs.Mutex.RLock()
	metrics := cs.Metrics
	cs.Mutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleMetrics handles metrics requests (exported for testing)
func (cs *CacheServer) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	cs.handleMetrics(w, r)
}

// HandleHealth handles health check requests (exported for testing)
func (cs *CacheServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	cs.handleHealth(w, r)
}

// handleHealth handles health check requests
func (cs *CacheServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"size":      cs.Metrics.Size,
		"entries":   cs.Metrics.Entries,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// updateMetrics updates cache metrics
func (cs *CacheServer) updateMetrics() {
	size, _ := cs.Storage.Size()
	entries, _ := cs.Storage.List()
	
	cs.Metrics.Size = size
	cs.Metrics.Entries = len(entries)
}

// startCleanupRoutine starts the periodic cleanup routine
func (cs *CacheServer) startCleanupRoutine() {
	ticker := time.NewTicker(cs.Config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			err := cs.Storage.Cleanup()
			if err != nil {
				log.Printf("Cleanup error: %v", err)
			} else {
				cs.Metrics.LastCleanup = time.Now()
				cs.updateMetrics()
			}
		case <-cs.shutdown:
			return
		}
	}
}