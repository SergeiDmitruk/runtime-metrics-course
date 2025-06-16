package storage

import (
	"context"
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
				storage.UpdateGauge(context.Background(), "temperature", 23.5)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics, _ := storage.GetMetrics(context.Background())
				if metrics.Gauges["temperature"] != 23.5 {
					t.Errorf("Expected gauge value 23.5, got %v", metrics.Gauges["temperature"])
				}
			},
		},
		{
			name: "Overwrite gauge",
			update: func(storage *MemStorage) {
				storage.UpdateGauge(context.Background(), "temperature", 23.5)
				storage.UpdateGauge(context.Background(), "temperature", 25.0)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics, _ := storage.GetMetrics(context.Background())
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
				storage.UpdateCounter(context.Background(), "requests", 10)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics, _ := storage.GetMetrics(context.Background())
				if metrics.Counters["requests"] != 10 {
					t.Errorf("Expected counter value 10, got %v", metrics.Counters["requests"])
				}
			},
		},
		{
			name: "Multiple updates",
			update: func(storage *MemStorage) {
				storage.UpdateCounter(context.Background(), "requests", 10)
				storage.UpdateCounter(context.Background(), "requests", 5)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics, _ := storage.GetMetrics(context.Background())
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
				metrics, _ := storage.GetMetrics(context.Background())
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
				storage.UpdateGauge(context.Background(), "temperature", 23.5)
				storage.UpdateCounter(context.Background(), "requests", 10)
			},
			verify: func(storage *MemStorage, t *testing.T) {
				metrics, _ := storage.GetMetrics(context.Background())
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

func BenchmarkUpdateGauge(b *testing.B) {
	storage := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.UpdateGauge(ctx, "test_gauge", float64(i))
	}
}

func BenchmarkUpdateCounter(b *testing.B) {
	storage := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.UpdateCounter(ctx, "test_counter", int64(i))
	}
}

func BenchmarkGetMetrics(b *testing.B) {
	storage := NewMemStorage()
	ctx := context.Background()

	for i := 0; i < 1000; i++ {
		_ = storage.UpdateGauge(ctx, "gauge_"+string(rune(i)), float64(i))
		_ = storage.UpdateCounter(ctx, "counter_"+string(rune(i)), int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetMetrics(ctx)
	}
}

func BenchmarkUpdateAll(b *testing.B) {
	storage := NewMemStorage()
	ctx := context.Background()

	metrics := make([]models.MetricJSON, 1000)
	for i := range metrics {
		if i%2 == 0 {
			value := float64(i)
			metrics[i] = models.MetricJSON{
				ID:    "gauge_" + string(rune(i)),
				MType: "gauge",
				Value: &value,
			}
		} else {
			delta := int64(i)
			metrics[i] = models.MetricJSON{
				ID:    "counter_" + string(rune(i)),
				MType: "counter",
				Delta: &delta,
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.UpdateAll(ctx, metrics)
	}
}

func BenchmarkConcurrentUpdates(b *testing.B) {
	storage := NewMemStorage()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				_ = storage.UpdateGauge(ctx, "concurrent_gauge", float64(i))
			} else {
				_ = storage.UpdateCounter(ctx, "concurrent_counter", int64(i))
			}
			i++
		}
	})
}
