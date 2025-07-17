// Package agent implements metrics collection and reporting functionality.
// It periodically collects system metrics and sends them to a specified server.
package agent

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"time"

	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/models"
)

// Config contains agent configuration parameters.
// Fields can be set via environment variables (see env tags).
type Config struct {
	Host           string         // Server address to report metrics to
	SecretKey      string         // Secret key for request signing
	CryptoKeyPath  string         // Path to public key
	PollInterval   time.Duration  // How often to collect metrics
	ReportInterval time.Duration  // How often to send metrics
	RateLimit      int            // Maximum concurrent requests
	PablicKey      *rsa.PublicKey // Public key for encrypt
	Ctx            context.Context
}

// Task represents a metric reporting task containing the metric to be sent.
type Task struct {
	Metric models.MetricJSON // Metric data in JSON format
}

// Global configuration instance
var cfg Config

// StartAgent initializes and runs the metrics collection and reporting agent.
// Parameters:
//   - conf: Agent configuration
//
// Returns:
//   - error: if initialization fails
//
// The agent runs two main loops:
//   - Poll loop: collects system metrics at regular intervals
//   - Report loop: sends collected metrics to server
//
// Example:
//
//	config := agent.Config{
//	    Host:           "localhost:8080",
//	    PollInterval:   2 * time.Second,
//	    ReportInterval: 10 * time.Second,
//	    RateLimit:      5,
//	}
//	if err := agent.StartAgent(config); err != nil {
//	    log.Fatal(err)
//	}
func StartAgent(conf Config) error {
	cfg = conf
	cfg.PablicKey = getPublicKey(cfg.CryptoKeyPath)
	// Initialize tickers for periodic operations
	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer func() {
		pollTicker.Stop()
		reportTicker.Stop()
	}()

	// Channel for metric reporting tasks
	taskChan := make(chan Task)
	defer close(taskChan)

	// Main agent loop
	for {
		select {
		case <-cfg.Ctx.Done():
			// Context cancelled, finish processing remaining tasks

			logger.Log.Info("Shutting down agent...")
			return nil
		case <-pollTicker.C:
			// Collect metrics in separate goroutines
			go CollectRuntimeMetrics(taskChan)
			go CollectGoupsutiMetrics(taskChan)
		case <-reportTicker.C:
			// Start workers to send metrics
			go startWorkerPool(cfg.Ctx, cfg.RateLimit, taskChan)
		}
	}
}

func getPublicKey(CryptKeyPath string) *rsa.PublicKey {
	keyBytes, err := os.ReadFile(CryptKeyPath)
	if err != nil {
		logger.Log.Sugar().Error(err)
		return nil
	}

	key, err := x509.ParsePKCS1PublicKey(keyBytes)
	if err != nil {
		logger.Log.Sugar().Error(err)
		return nil
	}

	return key
}
