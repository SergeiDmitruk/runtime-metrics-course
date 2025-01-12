package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/mocks"
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
				storage.On("UpdateCounter", "requests", int64(10)).Return(nil)
				storage.On("PrintMetrics").Return(nil)
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
			r.Route("/update", func(r chi.Router) {
				r.Post("/{metric_type}/{name}/{value}", UpdateHandler(storage))
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
				storage.On("GetGauges").Return(map[string]float64{
					"temperature": 25.5,
				})
			},
			expectedCode: http.StatusOK,
			expectedBody: "25.5",
		},
		{
			name:   "Valid counter metric",
			url:    "/value/counter/requests",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetCounters").Return(map[string]int64{
					"requests": 42,
				})
			},
			expectedCode: http.StatusOK,
			expectedBody: "42",
		},
		{
			name:   "Unknown gauge metric",
			url:    "/value/gauge/unknown",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetGauges").Return(map[string]float64{})
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Unknown metric\n",
		},
		{
			name:   "Unknown counter metric",
			url:    "/value/counter/unknown",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
				storage.On("GetCounters").Return(map[string]int64{})
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Unknown metric\n",
		},
		{
			name:   "Invalid metric type",
			url:    "/value/invalid/metric",
			method: http.MethodGet,
			setupMock: func(storage *mocks.StorageIface) {
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
			r.Route("/value", func(r chi.Router) {
				r.Get("/{metric_type}/{name}", GetMetricValueHandler(storage))
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
