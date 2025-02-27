package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/storage"
)

type Config struct {
	Host           string
	PollInterval   int `env:"REPORT_INTERVAL"`
	ReportInterval int `env:"POLL_INTERVAL"`
}

func main() {
	hostFlag := flag.String("a", "http://localhost:8080", "server config host:port")
	pollIntervalFlag := flag.Int("p", 2, "poll interval")
	reportIntervalFlag := flag.Int("r", 10, "report interval")

	flag.Parse()
	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}
	config := &Config{
		Host:           *hostFlag,
		PollInterval:   *pollIntervalFlag,
		ReportInterval: *reportIntervalFlag,
	}
	err := env.Parse(config)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if configEnv := os.Getenv("ADDRESS"); configEnv != "" {
		config.Host = configEnv
	}

	if !strings.Contains(config.Host, "http") {
		config.Host = "http://" + config.Host
	}
	sm, err := storage.NewStorageManager(nil)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	storage, err := sm.GetStorage()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	if err := agent.StartAgent(storage, config.Host, time.Duration(config.PollInterval)*time.Second, time.Duration(config.ReportInterval)*time.Second); err != nil {
		logger.Log.Fatal(err.Error())
	}
	logger.Log.Sugar().Info("agent start send to ", config.Host)
	select {}
}
