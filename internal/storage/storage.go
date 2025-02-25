package storage

import (
	"context"

	"github.com/runtime-metrics-course/internal/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=StorageIface --output=../mocks --outpkg=mocks --filename=storage_mock.go
type StorageIface interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	GetMetrics(ctx context.Context) (models.Metrics, error)
	Ping() error
}
