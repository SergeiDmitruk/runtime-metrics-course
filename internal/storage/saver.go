package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/utils"
)

type StorageWorker struct {
	interval    time.Duration
	filePath    string
	restore     bool
	storage     StorageIface
	stopChannel chan struct{}
}

func NewStorageWorker(interval time.Duration, filePath string, restore bool, storage StorageIface) *StorageWorker {
	return &StorageWorker{
		interval:    interval,
		filePath:    filePath,
		restore:     restore,
		storage:     storage,
		stopChannel: make(chan struct{}),
	}
}

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
			sw.storage.UpdateCounter(metric.ID, *metric.Delta)
		case metric.IsGauge():
			sw.storage.UpdateGauge(metric.ID, *metric.Value)
		}
	}
	log.Println("Метрики загружены из файла")
	return nil
}

func (sw *StorageWorker) SaveToFile() error {
	file, err := os.OpenFile(sw.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var data []*models.MetricJSON
	metrics := sw.storage.GetMetrics()

	for name, val := range metrics.Gauges {
		jm, err := utils.ParseMetricToJSON(models.Gauge, name, val)
		if err != nil {
			continue
		}
		data = append(data, jm)
	}

	for name, val := range metrics.Counters {
		jm, err := utils.ParseMetricToJSON(models.Counter, name, val)
		if err != nil {
			continue
		}
		data = append(data, jm)
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (sw *StorageWorker) SaverRun() {
	if err := sw.LoadFromFile(); err != nil {
		fmt.Println("Ошибка загрузки метрик:", err)
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
					fmt.Println("Ошибка сохранения метрик:", err)
				}
			case <-sw.stopChannel:
				return
			}
		}
	}()
}

func (sw *StorageWorker) SaverStop() {
	close(sw.stopChannel)
	if err := sw.SaveToFile(); err != nil {
		fmt.Println("Ошибка сохранения при завершении:", err)
	}
	fmt.Println("Метрики сохранены перед выходом")
}
