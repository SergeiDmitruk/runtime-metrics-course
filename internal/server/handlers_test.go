package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/mocks"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		method       string
		setupMock    func(storage *mocks.StorageIface)
		expectedCode int
	}{
		{
			name:   "Valid counter update",
			url:    "/update/counter/requests/10",
			method: http.MethodPost,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("UpdateCounter", mock.Anything, "requests", int64(10)).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "Missing metric value",
			url:    "/update/counter/requests/",
			method: http.MethodPost,
			setupMock: func(storage *mocks.StorageIface) {
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:   "Invalid metric type",
			url:    "/update/unknown_metric/something/123",
			method: http.MethodPost,
			setupMock: func(storage *mocks.StorageIface) {
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "Invalid gauge value",
			url:    "/update/gauge/temperature/not-a-number",
			method: http.MethodPost,
			setupMock: func(storage *mocks.StorageIface) {
			},
			expectedCode: http.StatusBadRequest,
		},

		{
			name:   "Invalid URL format",
			url:    "/update/",
			method: http.MethodPost,
			setupMock: func(storage *mocks.StorageIface) {
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewStorageIface(t)
			tt.setupMock(storage)

			r := chi.NewRouter()
			h := GetNewMetricsHandler(storage)
			r.Route("/update", func(r chi.Router) {
				r.Post("/{metric_type}/{name}/{value}", h.Update)
			})

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if status := w.Code; status != tt.expectedCode {
				t.Errorf("expected status %v, got %v", tt.expectedCode, status)
			}

			storage.AssertExpectations(t)
		})
	}
}

func TestGetMetricValue(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		method       string
		setupMock    func(storage *mocks.StorageIface)
		expectedCode int
		expectedBody string
	}{
		{
			name:   "Valid gauge metric",
			url:    "/value/gauge/temperature",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
					Gauges: models.Gauges{
						"temperature": 25.5,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "25.5",
		},
		{
			name:   "Valid counter metric",
			url:    "/value/counter/requests",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
					Counters: models.Counters{
						"requests": 42,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "42",
		},
		{
			name:   "Unknown gauge metric",
			url:    "/value/gauge/unknown",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
					Gauges: models.Gauges{},
				}, nil)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Unknown metric\n",
		},
		{
			name:   "Unknown counter metric",
			url:    "/value/counter/unknown",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
					Counters: models.Counters{},
				}, nil)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Unknown metric\n",
		},
		{
			name:   "Invalid metric type",
			url:    "/value/invalid/metric",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
					Counters: models.Counters{},
					Gauges:   models.Gauges{},
				}, nil)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Unknown metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewStorageIface(t)
			tt.setupMock(storage)

			r := chi.NewRouter()
			h := GetNewMetricsHandler(storage)
			r.Route("/value", func(r chi.Router) {
				r.Get("/{metric_type}/{name}", h.GetMetricValue)
			})

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if status := w.Code; status != tt.expectedCode {
				t.Errorf("expected status %v, got %v", tt.expectedCode, status)
			}

			if body := w.Body.String(); body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}

			storage.AssertExpectations(t)
		})
	}
}

func TestGetMetricValueJSONHandler(t *testing.T) {
	testValue := 25.5
	tests := []struct {
		name         string
		url          string
		method       string
		body         models.MetricJSON
		setupMock    func(storage *mocks.StorageIface)
		expectedCode int
		expectedBody models.MetricJSON
	}{
		{
			name:   "Valid gauge metric",
			url:    "/value/",
			method: http.MethodPost,
			body: models.MetricJSON{
				ID:    "temperature",
				MType: models.Gauge,
			},
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
					Gauges: models.Gauges{
						"temperature": 25.5,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: models.MetricJSON{
				ID:    "temperature",
				MType: models.Gauge,
				Value: &testValue,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewStorageIface(t)
			tt.setupMock(storage)

			r := chi.NewRouter()
			h := GetNewMetricsHandler(storage)
			r.Post("/value/", h.GetMetricValueJSON)
			testBody, err := json.Marshal(tt.body)
			require.NoError(t, err)
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(testBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if status := w.Code; status != tt.expectedCode {
				t.Errorf("expected status %v, got %v", tt.expectedCode, status)
			}
			var respStruct models.MetricJSON
			respBody, err := io.ReadAll(w.Body)
			if err != nil {
				t.Error(err)
			}

			if err = json.Unmarshal(respBody, &respStruct); err != nil {
				t.Error(err)
			}
			if !assert.Equal(t, tt.expectedBody, respStruct) {
				t.Error(*tt.expectedBody.Value, *respStruct.Value)
			}

			storage.AssertExpectations(t)
		})
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
				storage.On("UpdateAll", mock.Anything, mock.Anything).Return(errors.New("invalid metric"))

			},
			expectedCode: http.StatusInternalServerError,
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
func BenchmarkUpdateHandler(b *testing.B) {
	storage := mocks.NewStorageIface(b)
	storage.On("UpdateCounter", mock.Anything, "requests", int64(10)).Return(nil)

	r := chi.NewRouter()
	h := GetNewMetricsHandler(storage)
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metric_type}/{name}/{value}", h.Update)
	})

	req := httptest.NewRequest(http.MethodPost, "/update/counter/requests/10", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

func BenchmarkGetMetricValue(b *testing.B) {
	storage := mocks.NewStorageIface(b)
	storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
		Gauges: models.Gauges{
			"temperature": 25.5,
		},
	}, nil)

	r := chi.NewRouter()
	h := GetNewMetricsHandler(storage)
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metric_type}/{name}", h.GetMetricValue)
	})

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

func BenchmarkGetMetricValueJSON(b *testing.B) {
	storage := mocks.NewStorageIface(b)
	testValue := 25.5
	storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
		Gauges: models.Gauges{
			"temperature": testValue,
		},
	}, nil)

	r := chi.NewRouter()
	h := GetNewMetricsHandler(storage)
	r.Post("/value/", h.GetMetricValueJSON)

	metric := models.MetricJSON{
		ID:    "temperature",
		MType: models.Gauge,
	}
	testBody, _ := json.Marshal(metric)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(testBody))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateJSON(b *testing.B) {
	storage := mocks.NewStorageIface(b)
	storage.On("UpdateGauge", mock.Anything, "temperature", 25.5).Return(nil)
	storage.On("GetMetrics", mock.Anything).Return(models.Metrics{
		Gauges: models.Gauges{
			"temperature": 25.5,
		},
	}, nil)

	r := chi.NewRouter()
	h := GetNewMetricsHandler(storage)
	r.Post("/update/", h.UpdateJSON)

	metric := models.MetricJSON{
		ID:    "temperature",
		MType: models.Gauge,
		Value: pointerToFloat64(25.5),
	}
	testBody, _ := json.Marshal(metric)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(testBody))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateAll(b *testing.B) {
	storage := mocks.NewStorageIface(b)
	storage.On("UpdateAll", mock.Anything, mock.Anything).Return(nil)

	r := chi.NewRouter()
	h := GetNewMetricsHandler(storage)
	r.Post("/update/", h.UpdateAll)

	metrics := []models.MetricJSON{
		{ID: "temperature", MType: models.Gauge, Value: pointerToFloat64(25.5)},
		{ID: "humidity", MType: models.Gauge, Value: pointerToFloat64(60.0)},
	}
	testBody, _ := json.Marshal(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(testBody))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
