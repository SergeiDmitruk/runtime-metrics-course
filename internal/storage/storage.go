package storage

import (
	"github.com/runtime-metrics-course/internal/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=StorageIface --output=../mocks --outpkg=mocks --filename=storage_mock.go
type StorageIface interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetMetrics() models.Metrics
}
