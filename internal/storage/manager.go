package storage

import (
	"errors"
	"fmt"
	"log"
)

const ( //list of possible storage types
	RuntimeMemory = "mem_storage"
)

type StorageManager struct {
	*StorageWorker
	storage     StorageIface
	storageType string
}

var currentSM StorageManager

func GetStorageManager() *StorageManager {
	return &currentSM
}

func NewStorageManager(storageType string, workerCfg *WorkerCfg) (*StorageManager, error) {

	var err error
	switch storageType {
	case RuntimeMemory:
		currentSM.storage = NewMemStorage()

	default:
		return &currentSM, fmt.Errorf("unknown storge type - %s", storageType)
	}

	currentSM.storageType = storageType

	if workerCfg != nil {
		currentSM.StorageWorker = NewStorageWorker(workerCfg, currentSM.storage)
	}

	return &currentSM, err
}

func (m *StorageManager) GetStorage() (StorageIface, error) {
	log.Println("OK", m.storageType)
	if m.storage == nil {
		return nil, errors.New("storage is not initialized")
	}
	return m.storage, nil
}

func (m *StorageManager) GetStorageType() string {
	return m.storageType
}
