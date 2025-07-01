// Package storage provides abstractions and implementations for metrics storage.
//
// The package includes:
//   - StorageIface interface for metric storage operations
//   - In-memory implementation (MemStorage)
//   - PostgreSQL implementation (PgxStorage)
//   - Storage manager (StorageManager)
//   - Background worker for file persistence (StorageWorker)
//
// Core Components:
//
// Interfaces:
//   - StorageIface: Core interface for metric operations
//
// Storage Implementations:
//   - MemStorage: Thread-safe in-memory storage
//   - PgxStorage: PostgreSQL storage with caching
//
// Storage Management:
//   - StorageManager: Unified access point to storage
//   - StorageWorker: Periodic persistence for file storage
//
// Configuration:
//   - Cfg: Storage initialization settings
//
// Usage Example:
//
//	// Initialize storage manager
//	cfg := &storage.Cfg{
//	    Interval: time.Minute,
//	    FilePath: "/tmp/metrics.json",
//	    Restore:  true,
//	}
//	manager, _ := storage.NewStorageManager(cfg)
//
//	// Get storage instance
//	store, _ := manager.GetStorage()
//
//	// Work with metrics
//	store.UpdateGauge(ctx, "temperature", 23.5)
//	store.UpdateCounter(ctx, "requests", 1)
//	metrics, _ := store.GetMetrics(ctx)
//
// Concurrency Safety:
// All storage implementations guarantee thread-safe concurrent access.
//
// Storage Types:
// Constants defining supported storage backends:
//   - RuntimeMemory: In-memory storage
//   - PostgresDB: PostgreSQL storage
package storage
