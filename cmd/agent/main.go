package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/logger"
)

func main() {
	hostFlag := flag.String("a", "http://localhost:8080", "server config host:port")
	keyFlag := flag.String("k", "", "encrypt key")
	pollIntervalFlag := flag.Int("p", 2, "poll interval")
	reportIntervalFlag := flag.Int("r", 10, "report interval")
	rateLimit := flag.Int("l", 3, "rate limit")

	flag.Parse()
	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}
	config := agent.Config{
		Host:           *hostFlag,
		SecretKey:      *keyFlag,
		PollInterval:   time.Duration(*pollIntervalFlag) * time.Second,
		ReportInterval: time.Duration(*reportIntervalFlag) * time.Second,
		RateLimit:      *rateLimit,
	}
	err := env.Parse(&config)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if !strings.Contains(config.Host, "http") {
		config.Host = "http://" + config.Host
	}

	if err := agent.StartAgent(config); err != nil {
		logger.Log.Fatal(err.Error())
	}
	logger.Log.Sugar().Info("agent start send to ", config.Host)
	select {}
}
