package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/storage"
	"github.com/runtime-metrics-course/internal/templates"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MetricsHadler struct {
	storage storage.StorageIface
}

func GetNewMetricsHandler(storage storage.StorageIface) *MetricsHadler {
	return &MetricsHadler{storage: storage}
}

func (h *MetricsHadler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl := templates.GetMetricsTemplate()

	data, err := h.storage.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHadler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHadler) GetMetricValueJSON(w http.ResponseWriter, r *http.Request) {
	metric := &models.MetricJSON{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	metricName := metric.ID
	storageMetrics, err := h.storage.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch metric.MType {
	case Gauge:
		if val, ok := storageMetrics.Gauges[metricName]; ok {
			metric.Value = &val
		} else {
			http.Error(w, "Unknown metric", http.StatusNotFound)
			return
		}
	case Counter:
		if val, ok := storageMetrics.Counters[metricName]; ok {
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
	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHadler) Update(w http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "metric_type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(r.Context(), name, val)
	case Counter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(r.Context(), name, val)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h MetricsHadler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	metric := &models.MetricJSON{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, metric); err != nil {
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
		if metric.Value == nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(r.Context(), metric.ID, *metric.Value)

	case Counter:
		if metric.Delta == nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(r.Context(), metric.ID, *metric.Delta)

		if val, ok := storageMetrics.Counters[metric.ID]; ok {
			metric.Delta = &val
		}
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	respData, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respData)

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHadler) PingDBHandler(w http.ResponseWriter, r *http.Request) {

	if err := h.storage.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
