package storage

import (
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "Empty storage",
			want: &MemStorage{
				gauges:   make(map[string]float64),
				counters: make(map[string]int64),
			},
		},
		{
			name: "New instance",
			want: NewMemStorage(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			// Проверяем, что хранимые данные пусты
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
				if storage.gauges["temperature"] != 23.5 {
					t.Errorf("Expected gauge value 23.5, got %v", storage.gauges["temperature"])
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
				if storage.gauges["temperature"] != 25.0 {
					t.Errorf("Expected gauge value 25.0 after overwrite, got %v", storage.gauges["temperature"])
				}
			},
		},
		{
			name: "Multiple gauges",
			update: func(storage *MemStorage) {
				storage.UpdateGauge("temperature", 23.5)
				storage.UpdateGauge("humidity", 45.2)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				if storage.gauges["temperature"] != 23.5 {
					t.Errorf("Expected temperature value 23.5, got %v", storage.gauges["temperature"])
				}
				if storage.gauges["humidity"] != 45.2 {
					t.Errorf("Expected humidity value 45.2, got %v", storage.gauges["humidity"])
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
				if storage.counters["requests"] != 10 {
					t.Errorf("Expected counter value 10, got %v", storage.counters["requests"])
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
				if storage.counters["requests"] != 15 {
					t.Errorf("Expected counter value 15, got %v", storage.counters["requests"])
				}
			},
		},
		{
			name: "Different counters",
			update: func(storage *MemStorage) {
				storage.UpdateCounter("requests", 10)
				storage.UpdateCounter("errors", 3)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				if storage.counters["requests"] != 10 {
					t.Errorf("Expected counter value 10 for requests, got %v", storage.counters["requests"])
				}
				if storage.counters["errors"] != 3 {
					t.Errorf("Expected counter value 3 for errors, got %v", storage.counters["errors"])
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

func TestGetGauges(t *testing.T) {
	tests := []struct {
		name   string
		update func(*MemStorage)
		verify func(mem *MemStorage, t *testing.T)
	}{
		{
			name:   "Get empty gauges",
			update: func(storage *MemStorage) {},
			verify: func(storage *MemStorage, t *testing.T) {
				gauges := storage.GetGauges()
				if len(gauges) != 0 {
					t.Errorf("Expected empty gauges map, got %v", gauges)
				}
			},
		},
		{
			name: "Get gauges after update",
			update: func(storage *MemStorage) {
				storage.UpdateGauge("temperature", 23.5)
				storage.UpdateGauge("humidity", 45.2)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				gauges := storage.GetGauges()
				if len(gauges) != 2 {
					t.Errorf("Expected 2 gauges, got %v", len(gauges))
				}
				if gauges["temperature"] != 23.5 {
					t.Errorf("Expected temperature value 23.5, got %v", gauges["temperature"])
				}
				if gauges["humidity"] != 45.2 {
					t.Errorf("Expected humidity value 45.2, got %v", gauges["humidity"])
				}
			},
		},
		{
			name: "Get gauges with overwrite",
			update: func(storage *MemStorage) {
				storage.UpdateGauge("temperature", 23.5)
				storage.UpdateGauge("temperature", 25.0)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				gauges := storage.GetGauges()
				if gauges["temperature"] != 25.0 {
					t.Errorf("Expected temperature value 25.0 after overwrite, got %v", gauges["temperature"])
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

func TestGetCounters(t *testing.T) {
	tests := []struct {
		name   string
		update func(*MemStorage)
		verify func(mem *MemStorage, t *testing.T)
	}{
		{
			name:   "Get empty counters",
			update: func(storage *MemStorage) {},
			verify: func(storage *MemStorage, t *testing.T) {
				counters := storage.GetCounters()
				if len(counters) != 0 {
					t.Errorf("Expected empty counters map, got %v", counters)
				}
			},
		},
		{
			name: "Get counters after update",
			update: func(storage *MemStorage) {
				storage.UpdateCounter("requests", 10)
				storage.UpdateCounter("errors", 3)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				counters := storage.GetCounters()
				if len(counters) != 2 {
					t.Errorf("Expected 2 counters, got %v", len(counters))
				}
				if counters["requests"] != 10 {
					t.Errorf("Expected requests value 10, got %v", counters["requests"])
				}
				if counters["errors"] != 3 {
					t.Errorf("Expected errors value 3, got %v", counters["errors"])
				}
			},
		},
		{
			name: "Get counters with multiple updates",
			update: func(storage *MemStorage) {
				storage.UpdateCounter("requests", 10)
				storage.UpdateCounter("requests", 5)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				counters := storage.GetCounters()
				if counters["requests"] != 15 {
					t.Errorf("Expected requests value 15, got %v", counters["requests"])
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
