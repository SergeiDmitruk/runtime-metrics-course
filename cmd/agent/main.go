package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printBuildInfo() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}
func main() {
	printBuildInfo()
	hostFlag := flag.String("a", "http://localhost:8080", "server config host:port")
	keyFlag := flag.String("k", "", "encrypt key")
	cryptoPathFlag := flag.String("crypto-key", "", "путь к файлу с публичным ключем")
	pollIntervalFlag := flag.Int("p", 2, "poll interval")
	reportIntervalFlag := flag.Int("r", 10, "report interval")
	rateLimit := flag.Int("l", 10, "rate limit")

	flag.Parse()
	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}
	config := agent.Config{
		Host:           *hostFlag,
		SecretKey:      *keyFlag,
		CryptoKeyPath:  *cryptoPathFlag,
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
	_, err = url.Parse(config.Host)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if err := agent.StartAgent(config); err != nil {
		logger.Log.Fatal(err.Error())
	}
	logger.Log.Sugar().Info("agent start send to ", config.Host)
	select {}
}
