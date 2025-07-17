package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

type AgentConfig struct {
	Host           string        `json:"address"`
	SecretKey      string        `json:"key"`
	CryptoKeyPath  string        `json:"crypto_key"`
	PollInterval   time.Duration `json:"poll_interval"`
	ReportInterval time.Duration `json:"report_interval"`
	RateLimit      int           `json:"rate_limit"`
}

func printBuildInfo() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}

func main() {
	printBuildInfo()

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}

	if !strings.Contains(cfg.Host, "http") {
		cfg.Host = "http://" + cfg.Host
	}

	_, err = url.Parse(cfg.Host)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	agentConfig := agent.Config{
		Host:           cfg.Host,
		SecretKey:      cfg.SecretKey,
		CryptoKeyPath:  cfg.CryptoKeyPath,
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
		RateLimit:      cfg.RateLimit,
	}

	if err := agent.StartAgent(agentConfig); err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Sugar().Info("agent start send to ", agentConfig.Host)
	select {}
}

func LoadConfig() (*AgentConfig, error) {

	cfg := &AgentConfig{
		Host:           "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      10,
	}

	var configFile string

	if configFile == "" {
		configFile = os.Getenv("CONFIG")
	}

	if configFile != "" {
		fileData, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var fileCfg AgentConfig
		if err := json.Unmarshal(fileData, &fileCfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		if fileCfg.Host != "" {
			cfg.Host = fileCfg.Host
		}
		if fileCfg.PollInterval != 0 {
			cfg.PollInterval = fileCfg.PollInterval
		}
		if fileCfg.ReportInterval != 0 {
			cfg.ReportInterval = fileCfg.ReportInterval
		}
		if fileCfg.CryptoKeyPath != "" {
			cfg.CryptoKeyPath = fileCfg.CryptoKeyPath
		}
		if fileCfg.SecretKey != "" {
			cfg.SecretKey = fileCfg.SecretKey
		}
		if fileCfg.RateLimit != 0 {
			cfg.RateLimit = fileCfg.RateLimit
		}
	}

	flag.StringVar(&configFile, "c", "", "Path to config file")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.StringVar(&cfg.Host, "a", cfg.Host, "server config host:port")
	flag.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "encrypt key")
	flag.StringVar(&cfg.CryptoKeyPath, "crypto-key", cfg.CryptoKeyPath, "путь к файлу с публичным ключем")
	flag.DurationVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval")
	flag.DurationVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit")
	flag.Parse()

	if envHost := os.Getenv("ADDRESS"); envHost != "" {
		cfg.Host = envHost
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.SecretKey = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKeyPath = envCryptoKey
	}
	if envPollInt := os.Getenv("POLL_INTERVAL"); envPollInt != "" {
		if dur, err := time.ParseDuration(envPollInt); err == nil {
			cfg.PollInterval = dur
		}
	}
	if envReportInt := os.Getenv("REPORT_INTERVAL"); envReportInt != "" {
		if dur, err := time.ParseDuration(envReportInt); err == nil {
			cfg.ReportInterval = dur
		}
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		if val, err := strconv.Atoi(envRateLimit); err == nil {
			cfg.RateLimit = val
		}
	}

	return cfg, nil
}
