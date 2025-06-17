package storage

import (
	"context"

	"github.com/runtime-metrics-course/internal/models"
)

// StorageIface defines the interface for metrics storage operations.
//
// Implementations should provide thread-safe access to the underlying storage
// and handle all data persistence operations. The interface supports:
//   - Basic metric updates (gauges and counters)
//   - Bulk updates
//   - Metrics retrieval
//   - Storage health checks
//
//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=StorageIface --output=../mocks --outpkg=mocks --filename=storage_mock.go
type StorageIface interface {
	// UpdateGauge stores a gauge metric with the given name and value.
	UpdateGauge(ctx context.Context, name string, value float64) error

	// UpdateCounter stores a counter metric with the given name and value.
	// The implementation should handle incrementing existing values.
	UpdateCounter(ctx context.Context, name string, value int64) error

	// GetMetrics retrieves all stored metrics.
	// Returns Metrics struct containing all gauges and counters,
	GetMetrics(ctx context.Context) (models.Metrics, error)

	// UpdateAll performs a batch update of multiple metrics.
	// Should be atomic - either all updates succeed or none are applied.
	UpdateAll(ctx context.Context, metrics []models.MetricJSON) error

	// Ping checks the storage connectivity.
	// Returns nil if storage is accessible, error otherwise.
	Ping(ctx context.Context) error
}
