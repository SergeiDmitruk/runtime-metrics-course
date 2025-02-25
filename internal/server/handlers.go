package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/mocks"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/storage"
	"github.com/runtime-metrics-course/internal/templates"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MetricsHandler struct {
	storage storage.StorageIface
}

func GetNewMetricsHandler(storage storage.StorageIface) *MetricsHandler {
	return &MetricsHandler{storage: storage}
}

func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
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
}

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

func (h *MetricsHandler) GetMetricValueJSON(w http.ResponseWriter, r *http.Request) {
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
}

func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {

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
}

func (h MetricsHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
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

}

func (h *MetricsHandler) PingDBHandler(w http.ResponseWriter, r *http.Request) {

	if err := h.storage.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *MetricsHandler) UpdateAll(w http.ResponseWriter, r *http.Request) {
	var metrics []models.MetricJSON
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, &metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.storage.UpdateAll(r.Context(), metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func TestUpdateAllHandler(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		method       string
		body         []models.MetricJSON
		setupMock    func(storage *mocks.StorageIface)
		expectedCode int
	}{
		{
			name:   "Valid update of metrics",
			url:    "/update/",
			method: http.MethodPost,
			body: []models.MetricJSON{
				{ID: "temperature", MType: models.Gauge, Value: pointerToFloat64(25.5)},
				{ID: "humidity", MType: models.Gauge, Value: pointerToFloat64(60.0)},
			},
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("UpdateAll", mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "Invalid JSON body",
			url:    "/update/",
			method: http.MethodPost,
			body:   []models.MetricJSON{},
			setupMock: func(storage *mocks.StorageIface) {
				// No interactions with storage here, as it should not be reached
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "Internal server error on update",
			url:    "/update/",
			method: http.MethodPost,
			body: []models.MetricJSON{
				{ID: "temperature", MType: models.Gauge, Value: pointerToFloat64(25.5)},
			},
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("UpdateAll", mock.Anything, mock.Anything).Return(errors.New("internal error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewStorageIface(t)
			tt.setupMock(storage)

			r := chi.NewRouter()
			h := GetNewMetricsHandler(storage)
			r.Post("/update/", h.UpdateAll)

			testBody, err := json.Marshal(tt.body)
			require.NoError(t, err)
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(testBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if status := w.Code; status != tt.expectedCode {
				t.Errorf("expected status %v, got %v", tt.expectedCode, status)
			}

			storage.AssertExpectations(t)
		})
	}
}

func pointerToFloat64(v float64) *float64 {
	return &v
}
