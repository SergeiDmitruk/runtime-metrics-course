package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
			expectedCode: http.StatusBadRequest,
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
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mocks.NewStorageIface(t)
			tt.setupMock(storage)

			mux := http.NewServeMux()
			mux.Handle("/update/", http.StripPrefix("/update/", UpdateHandler(storage)))

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// Проверка статуса ответа
			if status := w.Code; status != tt.expectedCode {
				t.Errorf("expected status %v, got %v", tt.expectedCode, status)
			}

			// Убедиться, что все ожидания мока выполнены
			storage.AssertExpectations(t)
		})
	}
}
