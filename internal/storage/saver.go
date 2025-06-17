package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/models"
)

// StorageWorker handles periodic saving and restoring of metrics to/from file storage.
// It works in conjunction with any StorageIface implementation to provide persistence.
type StorageWorker struct {
	interval    time.Duration // How often to save metrics to file
	filePath    string        // Path to the storage file
	restore     bool          // Whether to load metrics on startup
	storage     StorageIface  // Underlying metrics storage implementation
	stopChannel chan struct{} // Channel for graceful shutdown
}

// NewStorageWorker creates a new StorageWorker instance with the given configuration.
// Parameters:
//   - cfg: Configuration containing file path, interval and restore settings
//   - storage: The storage implementation to use for metrics persistence
func NewStorageWorker(cfg *Cfg, storage StorageIface) *StorageWorker {
	return &StorageWorker{
		interval:    cfg.Interval,
		filePath:    cfg.FilePath,
		restore:     cfg.Restore,
		storage:     storage,
		stopChannel: make(chan struct{}),
	}
}

// LoadFromFile loads metrics from the configured file into storage.
// Only loads if restore flag is true. Returns nil if file doesn't exist.
// Returns:
//   - error: if file exists but cannot be read or contains invalid data
func (sw *StorageWorker) LoadFromFile() error {
	if !sw.restore {
		return nil
	}

	file, err := os.Open(sw.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var data []*models.MetricJSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	for _, metric := range data {
		switch {
		case metric.IsCounter():
			err := sw.storage.UpdateCounter(context.Background(), metric.ID, *metric.Delta)
			if err != nil {
				return err
			}
		case metric.IsGauge():
			err := sw.storage.UpdateGauge(context.Background(), metric.ID, *metric.Value)
			if err != nil {
				return err
			}
		}
	}
	logger.Log.Sugar().Infoln("Metrics loaded from file")
	return nil
}

// SaveToFile saves all current metrics to the configured file.
// Returns:
//   - error: if file cannot be created or metrics cannot be serialized
func (sw *StorageWorker) SaveToFile() error {
	file, err := os.OpenFile(sw.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var data []*models.MetricJSON
	metrics, err := sw.storage.GetMetrics(context.Background())
	if err != nil {
		return err
	}

	// Serialize gauge metrics
	for name, val := range metrics.Gauges {
		jm, err := models.MarshalMetricToJSON(models.Gauge, name, val)
		if err != nil {
			continue
		}
		data = append(data, jm)
	}

	// Serialize counter metrics
	for name, val := range metrics.Counters {
		jm, err := models.MarshalMetricToJSON(models.Counter, name, val)
		if err != nil {
			continue
		}
		data = append(data, jm)
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

// SaverRun starts the periodic save routine.
// First loads existing metrics if restore is enabled, then starts a goroutine
// that saves metrics at the configured interval until stopped.
func (sw *StorageWorker) SaverRun() {
	if err := sw.LoadFromFile(); err != nil {
		fmt.Println("Error loading metrics:", err)
	}

	if sw.interval == 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(sw.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sw.SaveToFile(); err != nil {
					fmt.Println("Error saving metrics:", err)
				}
			case <-sw.stopChannel:
				return
			}
		}
	}()
}

// SaverStop gracefully shuts down the saver routine.
// Stops periodic saves and performs one final save before exiting.
func (sw *StorageWorker) SaverStop() {
	close(sw.stopChannel)
	if err := sw.SaveToFile(); err != nil {
		fmt.Println("Error saving metrics on exit:", err)
	}
	fmt.Println("Metrics saved before shutdown")
}
