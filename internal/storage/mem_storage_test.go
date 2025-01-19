package storage

import (
	"testing"

	"github.com/runtime-metrics-course/internal/models"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "Empty storage",
			want: &MemStorage{
				gauges:   make(models.Gauges),
				counters: make(models.Counters),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			if len(storage.gauges) != 0 {
				t.Errorf("Expected empty gauges map, got %v", storage.gauges)
			}
			if len(storage.counters) != 0 {
				t.Errorf("Expected empty counters map, got %v", storage.counters)
			}
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	tests := []struct {
		name   string
		update func(*MemStorage)
		verify func(mem *MemStorage, t *testing.T)
	}{
		{
			name: "Single update",
			update: func(storage *MemStorage) {
				storage.UpdateGauge("temperature", 23.5)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics := storage.GetMetrics()
				if metrics.Gauges["temperature"] != 23.5 {
					t.Errorf("Expected gauge value 23.5, got %v", metrics.Gauges["temperature"])
				}
			},
		},
		{
			name: "Overwrite gauge",
			update: func(storage *MemStorage) {
				storage.UpdateGauge("temperature", 23.5)
				storage.UpdateGauge("temperature", 25.0)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics := storage.GetMetrics()
				if metrics.Gauges["temperature"] != 25.0 {
					t.Errorf("Expected gauge value 25.0 after overwrite, got %v", metrics.Gauges["temperature"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			tt.update(storage)
			tt.verify(storage, t)
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	tests := []struct {
		name   string
		update func(*MemStorage)
		verify func(mem *MemStorage, t *testing.T)
	}{
		{
			name: "Single update",
			update: func(storage *MemStorage) {
				storage.UpdateCounter("requests", 10)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics := storage.GetMetrics()
				if metrics.Counters["requests"] != 10 {
					t.Errorf("Expected counter value 10, got %v", metrics.Counters["requests"])
				}
			},
		},
		{
			name: "Multiple updates",
			update: func(storage *MemStorage) {
				storage.UpdateCounter("requests", 10)
				storage.UpdateCounter("requests", 5)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics := storage.GetMetrics()
				if metrics.Counters["requests"] != 15 {
					t.Errorf("Expected counter value 15, got %v", metrics.Counters["requests"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			tt.update(storage)
			tt.verify(storage, t)
		})
	}
}

func TestGetMetrics(t *testing.T) {
	tests := []struct {
		name   string
		update func(*MemStorage)
		verify func(mem *MemStorage, t *testing.T)
	}{
		{
			name:   "Get empty metrics",
			update: func(storage *MemStorage) {},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics := storage.GetMetrics()
				if len(metrics.Gauges) != 0 {
					t.Errorf("Expected empty gauges map, got %v", metrics.Gauges)
				}
				if len(metrics.Counters) != 0 {
					t.Errorf("Expected empty counters map, got %v", metrics.Counters)
				}
			},
		},
		{
			name: "Get metrics after updates",
			update: func(storage *MemStorage) {
				storage.UpdateGauge("temperature", 23.5)
				storage.UpdateCounter("requests", 10)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics := storage.GetMetrics()
				if metrics.Gauges["temperature"] != 23.5 {
					t.Errorf("Expected temperature gauge value 23.5, got %v", metrics.Gauges["temperature"])
				}
				if metrics.Counters["requests"] != 10 {
					t.Errorf("Expected requests counter value 10, got %v", metrics.Counters["requests"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			tt.update(storage)
			tt.verify(storage, t)
		})
	}
}
