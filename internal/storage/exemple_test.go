package storage_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/runtime-metrics-course/internal/storage"
)

// ExampleStorageManager demonstrates basic usage of the storage manager
// with different storage backends.
func ExampleStorageManager() {
	// Example 1: In-memory storage
	memCfg := &storage.Cfg{
		Interval: time.Minute,
		FilePath: "/tmp/metrics.json",
		Restore:  true,
	}
	memManager, _ := storage.NewStorageManager(memCfg)
	memStorage, _ := memManager.GetStorage()

	// Work with in-memory storage
	_ = memStorage.UpdateGauge(context.Background(), "temperature", 23.5)

	// Example 2: PostgreSQL storage
	db, _ := sql.Open("postgres", "connection_string")
	pgManager, _ := storage.NewStorageManager(&storage.Cfg{Conn: db})
	pgStorage, _ := pgManager.GetStorage()

	// Work with PostgreSQL storage
	_ = pgStorage.UpdateCounter(context.Background(), "requests", 1)

	// Output:
}

// ExampleMemStorage demonstrates direct usage of in-memory storage
func ExampleMemStorage() {
	store := storage.NewMemStorage()

	// Update metrics
	_ = store.UpdateGauge(context.Background(), "cpu_usage", 75.3)
	_ = store.UpdateCounter(context.Background(), "visits", 1)

	// Retrieve metrics
	metrics, _ := store.GetMetrics(context.Background())

	fmt.Printf("Gauges count: %d, Counters count: %d\n",
		len(metrics.Gauges), len(metrics.Counters))

	// Output:
	// Gauges count: 1, Counters count: 1
}

// ExamplePgxStorage demonstrates direct usage of PostgreSQL storage
func ExamplePgxStorage_usage() {
	// This example demonstrates PgxStorage usage pattern.
	// Real implementation requires working PostgreSQL connection.

	// Typical initialization:
	/*
	   db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname")
	   if err != nil {
	       log.Fatal(err)
	   }
	   defer db.Close()

	   store := storage.NewPgxStorage(db)
	*/

	// Example metric updates:
	/*
	   err = store.UpdateGauge(context.Background(), "memory_usage", 45.2)
	   if err != nil {
	       log.Printf("Failed to update gauge: %v", err)
	   }

	   err = store.UpdateCounter(context.Background(), "clicks", 5)
	   if err != nil {
	       log.Printf("Failed to update counter: %v", err)
	   }
	*/
	// (Example demonstrates usage pattern only)
	// Output:

}

// ExampleStorageWorker demonstrates the persistence worker
func ExampleStorageWorker() {
	cfg := &storage.Cfg{
		Interval: 5 * time.Second,
		FilePath: "/tmp/metrics_backup.json",
		Restore:  true,
	}

	worker := storage.NewStorageWorker(cfg, storage.NewMemStorage())

	// Start background persistence
	worker.SaverRun()

	// ... application runs ...

	// Stop gracefully
	worker.SaverStop()

	// Output: Metrics saved before shutdown
}
