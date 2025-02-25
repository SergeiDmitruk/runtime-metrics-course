package models

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Gauges map[string]float64
type Counters map[string]int64

type Metrics struct {
	Gauges   Gauges
	Counters Counters
}
type MetricJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *MetricJSON) IsCounter() bool {
	return m.MType == Counter
}

func (m *MetricJSON) IsGauge() bool {
	return m.MType == Gauge
}
