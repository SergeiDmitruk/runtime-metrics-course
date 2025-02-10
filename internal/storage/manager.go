package storage

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

const ( //list of possible storage types
	RuntimeMemory = "mem_storage"
	PostgresDB    = "pgx_storage"
)

type Cfg struct {
	Interval time.Duration
	FilePath string
	Restore  bool
	Conn     *sql.DB
}

type StorageManager struct {
	*StorageWorker
	storage     StorageIface
	storageType string
}

var currentSM StorageManager

func GetStorageManager() *StorageManager {
	return &currentSM
}

func NewStorageManager(cfg *Cfg) (*StorageManager, error) {

	var err error
	switch {
	case cfg.Conn != nil:
		currentSM.storage = NewPgxStorage(cfg.Conn)
		currentSM.storageType = PostgresDB

	default:
		currentSM.storage = NewMemStorage()
		currentSM.storageType = RuntimeMemory
	}

	if cfg != nil {
		currentSM.StorageWorker = NewStorageWorker(cfg, currentSM.storage)
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
