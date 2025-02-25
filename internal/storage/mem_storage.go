package storage

import (
	"errors"
	"sync"

	"github.com/runtime-metrics-course/internal/models"
)

type MemStorage struct {
	mu       sync.Mutex
	gauges   models.Gauges
	counters models.Counters
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(models.Gauges),
		counters: make(models.Counters),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

func (m *MemStorage) GetMetrics() models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	copyGauges := make(models.Gauges, len(m.gauges))
	for k, v := range m.gauges {
		copyGauges[k] = v
	}
	copyCounters := make(models.Counters, len(m.counters))
	for k, v := range m.counters {
		copyCounters[k] = v
	}

	return models.Metrics{Gauges: copyGauges, Counters: copyCounters}
}

func (m *MemStorage) Ping() error {
	return errors.New("db is not initialized")
}
