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
