package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/storage"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func GetMetricValue(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.ToLower(chi.URLParam(r, "name"))

		metricType := chi.URLParam(r, "metric_type")
		switch metricType {
		case Gauge:
			gauges := storage.GetGauges()
			if val, ok := gauges[name]; ok {
				w.Write([]byte(fmt.Sprintf("%f", val)))
			} else {
				http.Error(w, "Unknown metric", http.StatusNotFound)
				return
			}
		case Counter:
			counter := storage.GetCounters()
			if val, ok := counter[name]; ok {
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
		storage.PrintMetrics()

		w.WriteHeader(http.StatusOK)
	}
}
