package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/runtime-metrics-course/internal/compress"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/resilience"
	"github.com/runtime-metrics-course/internal/storage"
)

const batchSize = 100

func SendMetrics(storage storage.StorageIface, serverAddress string) error {

	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		logger.Log.Sugar().Errorln(err)
	}

	for name, value := range metrics.Gauges {
		baseURL, err := url.Parse(serverAddress)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}
		baseURL.Path += "/update/gauge/" + url.PathEscape(name) + "/" + url.PathEscape(fmt.Sprintf("%f", value))
		if err := sendRequest(client, baseURL.String(), nil); err != nil {
			logger.Log.Sugar().Errorf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	for name, value := range metrics.Counters {
		baseURL, err := url.Parse(serverAddress)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}
		baseURL.Path += "/update/counter/" + url.PathEscape(name) + "/" + url.PathEscape(fmt.Sprintf("%d", value))
		if err := sendRequest(client, baseURL.String(), nil); err != nil {
			logger.Log.Sugar().Errorf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	return nil
}

func SendMetricsJSON(storage storage.StorageIface, serverAddress string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		logger.Log.Sugar().Errorln(err)
	}

	for name, value := range metrics.Gauges {
		metric := &models.MetricJSON{
			MType: models.Gauge,
			ID:    name,
			Value: &value,
		}

		data, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Sugar().Errorf("Error marshal gauge %s: %v", name, err)
			return err
		}

		url := fmt.Sprintf("%s/update/", serverAddress)
		if err := sendRequest(client, url, data); err != nil {
			logger.Log.Sugar().Errorf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	for name, value := range metrics.Counters {
		metric := &models.MetricJSON{
			MType: models.Counter,
			ID:    name,
			Delta: &value,
		}

		data, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Sugar().Errorf("Error marshal gauge %s: %v", name, err)
			return err
		}

		url := fmt.Sprintf("%s/update/", serverAddress)
		if err := sendRequest(client, url, data); err != nil {
			logger.Log.Sugar().Errorf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	return nil
}

func SendAll(storage storage.StorageIface, serverAddress string) error {
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

	batch := make([]*models.MetricJSON, 0, batchSize)
	for i, metric := range allMetrics {
		batch = append(batch, metric)
		if len(batch) >= batchSize || i == len(allMetrics)-1 {
			data, err := json.Marshal(batch)
			if err != nil {
				return err
			}
			if err := sendRequest(client, baseURL.String(), data); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	return nil
}

func sendRequest(client *http.Client, url string, body []byte) error {
	operation := func() error {
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
		}
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()
		return nil
	}
	return resilience.Retry(context.Background(), operation)
}
