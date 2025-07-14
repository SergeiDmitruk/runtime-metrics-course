package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/runtime-metrics-course/internal/compress"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/middleware"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/resilience"
	"github.com/runtime-metrics-course/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	batchSize = 100 // Maximum number of metrics to send in a single batch
)

// startWorkerPool initializes a pool of workers to process metric sending tasks
// with rate limiting.
//
// Parameters:
//   - rateLimit: Maximum number of requests per second
//   - tasks: Channel receiving tasks to process
func startWorkerPool(rateLimit int, tasks <-chan Task) {
	limiter := rate.NewLimiter(rate.Limit(rateLimit), 1)
	go worker(tasks, limiter)
}

// worker processes metric sending tasks from the channel with rate limiting.
//
// Parameters:
//   - tasks: Channel to receive tasks from
//   - limiter: Rate limiter controlling request frequency
func worker(tasks <-chan Task, limiter *rate.Limiter) {
	client := &http.Client{Timeout: 5 * time.Second}
	for task := range tasks {
		// Wait for rate limiter allowance
		limiter.Wait(context.Background())

		data, _ := json.Marshal(task.Metric)
		logger.Log.Info("Worker sending metric", zap.String("metric", string(data)))

		baseURL, err := url.Parse(cfg.Host)
		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}
		baseURL.Path += "/update/"

		err = sendRequest(client, baseURL.String(), data, cfg.SecretKey)
		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}
	}
}

// SendMetrics sends all metrics to the server using URL-encoded format.
//
// Parameters:
//   - storage: Storage interface to get metrics from
//   - serverAddress: Target server URL
//   - key: Secret key for request signing
//
// Returns:
//   - error: if any send operation fails
func SendMetrics(storage storage.StorageIface, serverAddress, key string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		logger.Log.Sugar().Errorln(err)
		return err
	}

	// Send gauge metrics
	for name, value := range metrics.Gauges {
		baseURL, err := url.Parse(serverAddress)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}
		baseURL.Path = path.Join(baseURL.Path, "update", "gauge", url.PathEscape(name), url.PathEscape(fmt.Sprintf("%f", value)))

		if err := sendRequest(client, baseURL.String(), nil, key); err != nil {
			logger.Log.Sugar().Errorf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	// Send counter metrics
	for name, value := range metrics.Counters {
		baseURL, err := url.Parse(serverAddress)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}
		baseURL.Path = path.Join(baseURL.Path, "update", "counter", url.PathEscape(name), url.PathEscape(fmt.Sprintf("%d", value)))

		if err := sendRequest(client, baseURL.String(), nil, key); err != nil {
			logger.Log.Sugar().Errorf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	return nil
}

// SendMetricsJSON sends all metrics to the server using JSON format.
//
// Parameters:
//   - storage: Storage interface to get metrics from
//   - serverAddress: Target server URL
//   - key: Secret key for request signing
//
// Returns:
//   - error: if any send operation fails
func SendMetricsJSON(storage storage.StorageIface, serverAddress, key string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		logger.Log.Sugar().Errorln(err)
		return err
	}

	// Send gauge metrics
	for name, value := range metrics.Gauges {
		metric := &models.MetricJSON{
			MType: models.Gauge,
			ID:    name,
			Value: &value,
		}

		data, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Sugar().Errorf("Error marshaling gauge %s: %v", name, err)
			return err
		}

		url := fmt.Sprintf("%s/update/", serverAddress)
		if err := sendRequest(client, url, data, key); err != nil {
			logger.Log.Sugar().Errorf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	// Send counter metrics
	for name, value := range metrics.Counters {
		metric := &models.MetricJSON{
			MType: models.Counter,
			ID:    name,
			Delta: &value,
		}

		data, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Sugar().Errorf("Error marshaling counter %s: %v", name, err)
			return err
		}

		url := fmt.Sprintf("%s/update/", serverAddress)
		if err := sendRequest(client, url, data, key); err != nil {
			logger.Log.Sugar().Errorf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	return nil
}

// SendAll sends all metrics in batches using JSON format with retry logic.
//
// Parameters:
//   - storage: Storage interface to get metrics from
//   - serverAddress: Target server URL
//   - key: Secret key for request signing
//
// Returns:
//   - error: if any send operation fails after retries
func SendAll(storage storage.StorageIface, serverAddress, key string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL, err := url.Parse(serverAddress)
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	baseURL.Path += "/updates/"

	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		return err
	}

	// Prepare all metrics for batching
	allMetrics := make([]*models.MetricJSON, 0)
	for key, v := range metrics.Gauges {
		if m, err := models.MarshalMetricToJSON(models.Gauge, key, v); err == nil {
			allMetrics = append(allMetrics, m)
		}
	}
	for key, v := range metrics.Counters {
		if m, err := models.MarshalMetricToJSON(models.Counter, key, v); err == nil {
			allMetrics = append(allMetrics, m)
		}
	}

	// Send metrics in batches
	batch := make([]*models.MetricJSON, 0, batchSize)
	for i, metric := range allMetrics {
		batch = append(batch, metric)
		if len(batch) >= batchSize || i == len(allMetrics)-1 {
			data, err := json.Marshal(batch)
			if err != nil {
				return err
			}
			if err := resilience.Retry(context.Background(), func() error {
				return sendRequest(client, baseURL.String(), data, key)
			}); err != nil {
				return err
			}
			batch = batch[:0] // Reset batch
		}
	}

	return nil
}

// sendRequest sends an HTTP request with compression and optional signing.
//
// Parameters:
//   - client: HTTP client to use
//   - url: Target URL
//   - body: Request body (nil for empty)
//   - key: Secret key for signing (empty for no signing)
//
// Returns:
//   - error: if request fails
func sendRequest(client *http.Client, url string, body []byte, key string) error {
	cbody, err := compress.CompressGzip(body)
	if err != nil {
		return fmt.Errorf("failed to compress request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(cbody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		if key != "" {
			req.Header.Set("HashSHA256", middleware.HmacSHA256(body, []byte(key)))
		}
	}
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
