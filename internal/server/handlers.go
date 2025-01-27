package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/storage"
	"github.com/runtime-metrics-course/internal/templates"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func GetMetricsHandler(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := templates.GetMetricsTemplate()
		if err != nil {
			http.Error(w, "Failed to load template", http.StatusInternalServerError)
			return
		}
		data := storage.GetMetrics()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func GetMetricValueHandler(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.ToLower(chi.URLParam(r, "name"))

		metricType := chi.URLParam(r, "metric_type")
		metrics := storage.GetMetrics()
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
}
func GetMetricValueJSONHandler(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := &models.MetricJSON{}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(data, metric); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		metricName := strings.ToLower(metric.ID)
		storageMetrics := storage.GetMetrics()
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
}
func UpdateHandler(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
			storage.UpdateGauge(strings.ToLower(name), val)
		case Counter:
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(strings.ToLower(name), val)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func UpdateJSONHandler(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := &models.MetricJSON{}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(data, metric); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		storageMetrics := storage.GetMetrics()
		switch metric.MType {
		case Gauge:
			if metric.Value == nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(strings.ToLower(metric.ID), *metric.Value)

		case Counter:
			if metric.Delta == nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(strings.ToLower(metric.ID), *metric.Delta)

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
}
