package storage

import (
	"database/sql"
	"errors"
	"time"
)

// Storage types constants
const (
	RuntimeMemory = "mem_storage" // In-memory storage type
	PostgresDB    = "pgx_storage" // PostgreSQL storage type
)

// Cfg contains configuration for storage management
type Cfg struct {
	Conn     *sql.DB       // Database connection (for PostgreSQL storage)
	Interval time.Duration // Interval for periodic saves (for memory storage)
	FilePath string        // File path for persistence (for memory storage)
	Restore  bool          // Whether to restore from file on startup

}

// StorageManager manages the application's storage backend.
// It provides a unified interface to different storage implementations
// and handles initialization and lifecycle management.
type StorageManager struct {
	*StorageWorker              // Embedded worker for file storage operations
	storage        StorageIface // Current storage implementation
	storageType    string       // Type of active storage
}

// Package-level singleton instance
var currentSM StorageManager

// GetStorageManager returns the current storage manager instance.
// Implements the singleton pattern.
func GetStorageManager() *StorageManager {
	return &currentSM
}

// NewStorageManager initializes the storage system based on configuration.
// Parameters:
//   - cfg: Configuration for storage setup. If nil, defaults to memory storage.
//
// Returns:
//   - *StorageManager: initialized storage manager
//   - error: initialization error if any
//
// Storage selection logic:
//   - Uses PostgreSQL if connection is provided in config
//   - Falls back to in-memory storage otherwise
func NewStorageManager(cfg *Cfg) (*StorageManager, error) {
	var err error

	switch {
	case cfg != nil && cfg.Conn != nil:
		currentSM.storage = NewPgxStorage(cfg.Conn)
		currentSM.storageType = PostgresDB
	default:
		currentSM.storage = NewMemStorage()
		currentSM.storageType = RuntimeMemory
	}

	// Initialize storage worker if configuration provided
	if cfg != nil {
		currentSM.StorageWorker = NewStorageWorker(cfg, currentSM.storage)
	}

	return &currentSM, err
}

// GetStorage returns the current storage implementation.
// Returns:
//   - StorageIface: active storage implementation
//   - error: if storage is not initialized
func (m *StorageManager) GetStorage() (StorageIface, error) {
	if m.storage == nil {
		return nil, errors.New("storage is not initialized")
	}
	return m.storage, nil
}

// GetStorageType returns the type of currently active storage.
// Returns one of the storage type constants (RuntimeMemory or PostgresDB).
func (m *StorageManager) GetStorageType() string {
	return m.storageType
}

// SaverRun starts the periodic save routine for file storage.
// No-op for non-memory storage types.
func (m *StorageManager) SaverRun() {
	if m.storageType != RuntimeMemory {
		return
	}
	m.StorageWorker.SaverRun()
}

// SaverStop gracefully stops the save routine for file storage.
// No-op for non-memory storage types.
func (m *StorageManager) SaverStop() {
	if m.storageType != RuntimeMemory {
		return
	}
	m.StorageWorker.SaverStop()
}
