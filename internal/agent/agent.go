package agent

import (
	"time"

	"github.com/runtime-metrics-course/internal/models"
)

type Config struct {
	Host           string
	SecretKey      string
	PollInterval   time.Duration
	ReportInterval time.Duration
	RateLimit      int
}

var cfg Config

type Task struct {
	Metric models.MetricJSON
}

func StartAgent(conf Config) error {
	cfg = conf
	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)
	taskChan := make(chan Task)
	for {
		select {
		case <-pollTicker.C:
			go CollectRuntimeMetrics(taskChan)
			go CollectGoupsutiMetrics(taskChan)
		case <-reportTicker.C:

		}
	}

	// go func() {
	// 	for range pollTicker.C {
	// 		go CollectRuntimeMetrics(taskChan)
	// 	}
	// }()
	// go func() {
	// 	for range reportTicker.C {
	// 		//SendMetrics(storage, address)
	// 		//SendMetricsJSON(storage, address)
	// 		SendAll(storage, address, key)
	// 	}

	// }()
	return nil
}
