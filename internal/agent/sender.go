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

func SendMetrics(storage storage.StorageIface, serverAddress string) error {

	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		logger.Log.Sugar().Errorln(err)
	}

	for name, value := range metrics.Gauges {
		baseUrl, err := url.Parse(serverAddress)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}
		baseUrl.Path += "/update/gauge/" + url.PathEscape(name) + "/" + url.PathEscape(fmt.Sprintf("%f", value))
		if err := sendRequest(client, baseUrl.String(), nil); err != nil {
			logger.Log.Sugar().Errorf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	for name, value := range metrics.Counters {
		baseUrl, err := url.Parse(serverAddress)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}
		baseUrl.Path += "/update/counter/" + url.PathEscape(name) + "/" + url.PathEscape(fmt.Sprintf("%d", value))
		if err := sendRequest(client, baseUrl.String(), nil); err != nil {
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
	baseUrl, err := url.Parse(serverAddress)
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	baseUrl.Path += "/updates/"
	allMetrics := make([]*models.MetricJSON, 0)
	batch := make([]*models.MetricJSON, 0, 100)
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		return err
	}
	for key, v := range metrics.Gauges {
		m, err := models.MarshalMetricToJSON(models.Gauge, key, v)
		if err == nil {
			allMetrics = append(allMetrics, m)
		}
	}

	for key, v := range metrics.Counters {
		m, err := models.MarshalMetricToJSON(models.Counter, key, v)
		if err == nil {
			allMetrics = append(allMetrics, m)
		}
	}

	for i := 0; i < len(allMetrics); i++ {
		batch = append(batch, allMetrics[i])
		if len(batch) == cap(batch) || i == len(allMetrics)-1 {

			data, err := json.Marshal(batch)
			if err != nil {
				return err
			}

			err = sendRequest(client, baseUrl.String(), data)
			if err != nil {
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
		} else {
			req.Header.Set("Content-Type", "text/plain")
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
