package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/server"
	"github.com/runtime-metrics-course/internal/storage"
)

// ExampleMetricsHandler_Update demonstrates how to update metrics via URL parameters
func ExampleMetricsHandler_Update() {
	storage := storage.NewMemStorage()
	h := server.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Post("/update/{metric_type}/{name}/{value}", h.Update)

	// Update a counter metric
	req := httptest.NewRequest("POST", "/update/counter/requests/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Update a gauge metric
	req = httptest.NewRequest("POST", "/update/gauge/temperature/22.5", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

// ExampleMetricsHandler_GetMetricValue shows how to retrieve a metric value
func ExampleMetricsHandler_GetMetricValue() {
	storage := storage.NewMemStorage()
	_ = storage.UpdateGauge(context.Background(), "temp", 23.5)
	h := server.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Get("/value/{metric_type}/{name}", h.GetMetricValue)

	req := httptest.NewRequest("GET", "/value/gauge/temp", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Response body will contain the metric value (23.5 in this case)
}

// ExampleMetricsHandler_UpdateJSON demonstrates JSON-based metric updates
func ExampleMetricsHandler_UpdateJSON() {
	storage := storage.NewMemStorage()
	h := server.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Post("/update/", h.UpdateJSON)

	metric := models.MetricJSON{
		ID:    "temperature",
		MType: "gauge",
		Value: new(float64),
	}
	*metric.Value = 22.5
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Response will contain the updated metric in JSON format
}

// ExampleMetricsHandler_GetMetricValueJSON shows JSON metric retrieval
func ExampleMetricsHandler_GetMetricValueJSON() {
	storage := storage.NewMemStorage()
	_ = storage.UpdateCounter(context.Background(), "requests", 42)
	h := server.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Post("/value/", h.GetMetricValueJSON)

	metric := models.MetricJSON{
		ID:    "requests",
		MType: "counter",
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/value/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Response will contain metric data in JSON format
}

// ExampleMetricsHandler_UpdateAll demonstrates batch metric updates
func ExampleMetricsHandler_UpdateAll() {
	storage := storage.NewMemStorage()
	h := server.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Post("/updates/", h.UpdateAll)

	metrics := []models.MetricJSON{
		{
			ID:    "requests",
			MType: "counter",
			Delta: new(int64),
		},
		{
			ID:    "temperature",
			MType: "gauge",
			Value: new(float64),
		},
	}
	*metrics[0].Delta = 10
	*metrics[1].Value = 23.5

	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

// ExampleMetricsHandler_PingDBHandler shows database connectivity check
func ExampleMetricsHandler_PingDBHandler() {
	storage := storage.NewMemStorage()
	h := server.NewMetricsHandler(storage)

	r := chi.NewRouter()
	r.Get("/ping", h.PingDBHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}
