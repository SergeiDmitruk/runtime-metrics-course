package storage

import (
	"fmt"
	"sync"
)

type MemStorage struct {
	mu       sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
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

func (m *MemStorage) GetGauges() map[string]float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	copyGauges := make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		copyGauges[k] = v
	}
	return copyGauges
}
func (m *MemStorage) GetCounters() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	copyGauges := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		copyGauges[k] = v
	}
	return copyGauges
}

func (m *MemStorage) PrintMetrics() { // test stdout
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("------Metrics-------")
	for name, val := range m.gauges {
		fmt.Println(name, val)
	}
	for name, val := range m.counters {
		fmt.Println(name, val)
	}
	fmt.Println("--------------------")
}
