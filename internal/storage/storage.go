package storage

import (
	"errors"
	"fmt"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)
const ( //list of possible db types
	RuntimeMemory = "mem_storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name=StorageIface --output=../mocks --outpkg=mocks
type StorageIface interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauges() map[string]float64
	GetCounters() map[string]int64
	PrintMetrics()
}

var currentStorage StorageIface

func InitStorage(storageType string) error {
	switch storageType {
	case RuntimeMemory:
		currentStorage = NewMemStorage()
		return nil
	default:
		return fmt.Errorf("unknown storge type - %s", storageType)
	}
}

func GetStorage() (StorageIface, error) {
	if currentStorage == nil {
		return nil, errors.New("storage is not initialized")
	}
	return currentStorage, nil
}
