// Package models defines the core data structures and operations for metrics handling.
// It includes types for both in-memory and JSON representations of metrics.
package models

import (
	"errors"

	"github.com/runtime-metrics-course/internal/logger"
)

// Metric type constants
const (
	Gauge   = "gauge"   // Gauge metric type (float values that can go up/down)
	Counter = "counter" // Counter metric type (monotonically increasing values)
)

// Gauges represents a collection of gauge metrics
type Gauges map[string]float64

// Counters represents a collection of counter metrics
type Counters map[string]int64

// Metrics aggregates all collected metrics
type Metrics struct {
	Gauges   Gauges   // Map of gauge metrics
	Counters Counters // Map of counter metrics
}

// MetricJSON represents a metric in JSON format for API communication
type MetricJSON struct {
	Delta *int64   `json:"delta,omitempty"` // Value for counter metrics
	Value *float64 `json:"value,omitempty"` // Value for gauge metrics
	ID    string   `json:"id"`              // Metric name
	MType string   `json:"type"`            // Metric type (gauge or counter)

}

// IsCounter checks if the metric is a counter type
func (m *MetricJSON) IsCounter() bool {
	return m.MType == Counter
}

// IsGauge checks if the metric is a gauge type
func (m *MetricJSON) IsGauge() bool {
	return m.MType == Gauge
}

// MarshalMetricToJSON creates a MetricJSON from raw values
// Parameters:
//   - mType: Metric type ("gauge" or "counter")
//   - name: Metric name
//   - val: Metric value (float64 for gauge, int64 for counter)
//
// Returns:
//   - *MetricJSON: constructed metric
//   - error: if type conversion fails
func MarshalMetricToJSON(mType, name string, val interface{}) (*MetricJSON, error) {
	metric := MetricJSON{
		ID:    name,
		MType: mType,
	}

	switch mType {
	case Counter:
		valInt, ok := val.(int64)
		if !ok {
			logger.Log.Error("parse error")
			return &metric, errors.New("invalid counter value type")
		}
		metric.Delta = &valInt

	case Gauge:
		valFl, ok := val.(float64)
		if !ok {
			logger.Log.Error("parse error")
			return &metric, errors.New("invalid gauge value type")
		}
		metric.Value = &valFl

	default:
		logger.Log.Error("parse error")
		return &metric, errors.New("invalid metric type")
	}

	return &metric, nil
}
