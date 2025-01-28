package agent

import (
	"log"
	"time"

	"github.com/runtime-metrics-course/internal/storage"
)

func StartAgent(storage storage.StorageIface, address string, pollInterval, reportInterval time.Duration) error {

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)
	go func() {
		for range pollTicker.C {
			CollectRuntimeMetrics(storage)
		}
	}()
	go func() {
		for range reportTicker.C {
			log.Println("------Sending metrics------")
			SendMetrics(storage, address)
			SendMetricsJSON(storage, address)
			log.Println("------Metrics sent successfully------")
		}

	}()
	return nil
}
