package models

import (
	"errors"

	"github.com/runtime-metrics-course/internal/logger"
)

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
			return &metric, errors.New("parse error")
		}
		metric.Delta = &valInt
	case Gauge:

		valFl, ok := val.(float64)
		if !ok {
			logger.Log.Error("parse error")
			return &metric, errors.New("parse error")
		}
		metric.Value = &valFl
	default:
		logger.Log.Error("parse error")
		return &metric, errors.New("parse error")
	}
	return &metric, nil
}
