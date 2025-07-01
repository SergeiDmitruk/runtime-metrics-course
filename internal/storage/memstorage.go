package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/runtime-metrics-course/internal/models"
)

// MemStorage implements StorageIface using in-memory storage with mutex protection.
// It provides thread-safe storage for gauge and counter metrics.
type MemStorage struct {
	mu       sync.Mutex      // Mutex to protect concurrent access
	gauges   models.Gauges   // Map for storing gauge metrics
	counters models.Counters // Map for storing counter metrics
}

// NewMemStorage creates a new initialized MemStorage instance.
// Returns:
//   - *MemStorage: ready-to-use in-memory storage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(models.Gauges),
		counters: make(models.Counters),
	}
}

// UpdateGauge stores or updates a gauge metric value.
// Implements StorageIface.UpdateGauge.
func (m *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
	return nil
}

// UpdateCounter increments a counter metric by the specified value.
// Implements StorageIface.UpdateCounter.
func (m *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
	return nil
}

// GetMetrics returns a snapshot of all stored metrics.
// Implements StorageIface.GetMetrics.
// Returns:
//   - models.Metrics: copy of all stored metrics
//   - error: always nil in this implementation
func (m *MemStorage) GetMetrics(ctx context.Context) (models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create defensive copies to prevent external modification
	copyGauges := make(models.Gauges, len(m.gauges))
	for k, v := range m.gauges {
		copyGauges[k] = v
	}

	copyCounters := make(models.Counters, len(m.counters))
	for k, v := range m.counters {
		copyCounters[k] = v
	}

	return models.Metrics{
		Gauges:   copyGauges,
		Counters: copyCounters,
	}, nil
}

// Ping always returns nil as in-memory storage is always available.
// Implements StorageIface.Ping.
func (m *MemStorage) Ping(ctx context.Context) error {
	return nil
}

// UpdateAll performs batch updates of multiple metrics atomically.
// Implements StorageIface.UpdateAll.
// Returns:
//   - error: joined error for all failed updates or nil if all succeeded
func (m *MemStorage) UpdateAll(ctx context.Context, metrics []models.MetricJSON) error {
	var errs []error
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range metrics {
		switch {
		case metric.IsCounter() && metric.Delta != nil:
			m.counters[metric.ID] += *metric.Delta
		case metric.IsGauge() && metric.Value != nil:
			m.gauges[metric.ID] = *metric.Value
		default:
			errs = append(errs, fmt.Errorf("%s: invalid metric type or value", metric.ID))
		}
	}

	return errors.Join(errs...)
}
