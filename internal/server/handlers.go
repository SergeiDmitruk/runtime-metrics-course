package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/resilience"
	"github.com/runtime-metrics-course/internal/storage"
	"github.com/runtime-metrics-course/internal/templates"
)

const (
	// Gauge represents a gauge metric type that can go up and down
	Gauge = "gauge"
	// Counter represents a counter metric type that only increments
	Counter = "counter"
)

// MetricsHandler handles all metrics-related HTTP requests
type MetricsHandler struct {
	storage storage.StorageIface // Storage interface for metrics persistence
}

// NewMetricsHandler creates a new MetricsHandler instance
func NewMetricsHandler(storage storage.StorageIface) *MetricsHandler {
	return &MetricsHandler{storage: storage}
}

// GetMetrics handles GET / - returns HTML page with all metrics
// Responses:
//   - 200: HTML page with metrics
//   - 500: Internal server error
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl := templates.GetMetricsTemplate()

	data, err := h.storage.GetMetrics(r.Context())
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// GetMetricValue handles GET /value/{metric_type}/{name} - returns plaintext metric value
// Responses:
//   - 200: Metric value as plaintext
//   - 400: Invalid metric type
//   - 404: Metric not found
//   - 500: Internal server error
func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	metricType := chi.URLParam(r, "metric_type")

	metrics, err := h.storage.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch metricType {
	case Gauge:
		if val, ok := metrics.Gauges[name]; ok {
			w.Write([]byte(strconv.FormatFloat(val, 'f', -1, 64)))
		} else {
			http.Error(w, "Unknown metric", http.StatusNotFound)
			return
		}
	case Counter:
		if val, ok := metrics.Counters[name]; ok {
			w.Write([]byte(fmt.Sprintf("%d", val)))
		} else {
			http.Error(w, "Unknown metric", http.StatusNotFound)
			return
		}
	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}
}

// GetMetricValueJSON handles POST /value/ - returns metric value in JSON format
// Responses:
//   - 200: JSON response with metric value
//   - 400: Invalid JSON or metric type
//   - 404: Metric not found
//   - 500: Internal server error
func (h *MetricsHandler) GetMetricValueJSON(w http.ResponseWriter, r *http.Request) {
	metric := &models.MetricJSON{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(data, metric); err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	storageMetrics, err := h.storage.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch metric.MType {
	case Gauge:
		if val, ok := storageMetrics.Gauges[metric.ID]; ok {
			metric.Value = &val
		} else {
			http.Error(w, "Unknown metric", http.StatusNotFound)
			return
		}
	case Counter:
		if val, ok := storageMetrics.Counters[metric.ID]; ok {
			metric.Delta = &val
		} else {
			http.Error(w, "Unknown metric", http.StatusNotFound)
			return
		}
	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	respData, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respData)
}

// Update handles POST /update/{metric_type}/{name}/{value} - updates metric via URL params
// Responses:
//   - 200: Metric updated successfully
//   - 400: Invalid metric type or value
//   - 500: Internal server error
func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			logger.Log.Error("Invalid gauge value")
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		err = resilience.Retry(r.Context(), func() error {
			return h.storage.UpdateGauge(r.Context(), name, val)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Counter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			logger.Log.Error("Invalid counter value")
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		err = resilience.Retry(r.Context(), func() error {
			return h.storage.UpdateCounter(r.Context(), name, val)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		logger.Log.Error("Invalid metric type")
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
}

// UpdateJSON handles POST /update/ - updates metric via JSON body
// Responses:
//   - 200: Metric updated successfully (returns updated metric in JSON)
//   - 400: Invalid JSON, metric type or value
//   - 500: Internal server error
func (h MetricsHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	metric := &models.MetricJSON{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(data, metric); err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case Gauge:
		if metric.Value == nil {
			logger.Log.Error("Invalid gauge value")
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		err = resilience.Retry(r.Context(), func() error {
			return h.storage.UpdateGauge(r.Context(), metric.ID, *metric.Value)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Counter:
		if metric.Delta == nil {
			logger.Log.Error("Invalid counter value")
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		err = resilience.Retry(r.Context(), func() error {
			return h.storage.UpdateCounter(r.Context(), metric.ID, *metric.Delta)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		logger.Log.Error("Invalid metric type")
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	respData, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, "", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respData)
}

// PingDBHandler handles GET /ping - checks database connectivity
// Responses:
//   - 200: Database is reachable
//   - 500: Database connection error
func (h *MetricsHandler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.storage.Ping(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateAll handles POST /updates/ - batch updates multiple metrics
// Responses:
//   - 200: Metrics updated successfully
//   - 400: Invalid JSON input
//   - 500: Internal server error
func (h *MetricsHandler) UpdateAll(w http.ResponseWriter, r *http.Request) {
	var metrics []models.MetricJSON
	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(data, &metrics); err != nil {
		logger.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	operation := func() error {
		return h.storage.UpdateAll(r.Context(), metrics)
	}

	if err := resilience.Retry(r.Context(), operation); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
