package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/runtime-metrics-course/internal/storage"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func UpdateHandler(storage storage.StorageIface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		//path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(r.URL.Path, "/")

		if len(parts) != 3 {
			if len(parts) == 2 && parts[1] == "" {
				http.Error(w, "Metric name is required", http.StatusNotFound)
				return
			}
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		metricType, name, value := parts[0], parts[1], parts[2]

		switch metricType {
		case Gauge:
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(name, val)
		case Counter:
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(name, val)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		storage.PrintMetrics()

		w.WriteHeader(http.StatusOK)
	}
}
