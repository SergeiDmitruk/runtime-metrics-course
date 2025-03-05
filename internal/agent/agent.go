package agent

import (
	"time"

	"github.com/runtime-metrics-course/internal/models"
)

type Config struct {
	Host           string        `env:"ADDRESS"`
	SecretKey      string        `env:"KEY"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	RateLimit      int           `env:"RATE_LIMIT"`
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
	defer close(taskChan)
	for {
		select {
		case <-pollTicker.C:
			go CollectRuntimeMetrics(taskChan)
			go CollectGoupsutiMetrics(taskChan)
		case <-reportTicker.C:
			go startWorkerPool(cfg.RateLimit, taskChan)
		}
	}

}
