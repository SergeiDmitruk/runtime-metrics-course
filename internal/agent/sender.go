package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/storage"
	"github.com/runtime-metrics-course/internal/utils"
)

func SendMetrics(storage storage.StorageIface, serverAddress string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		log.Println(err)
	}

	for name, value := range metrics.Gauges {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", serverAddress, name, value)
		if err := sendRequest(client, url, nil); err != nil {
			log.Printf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	for name, value := range metrics.Counters {
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverAddress, name, value)
		if err := sendRequest(client, url, nil); err != nil {
			log.Printf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	return nil
}

func SendMetricsJSON(storage storage.StorageIface, serverAddress string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		log.Println(err)
	}

	for name, value := range metrics.Gauges {
		metric := &models.MetricJSON{
			MType: models.Gauge,
			ID:    name,
			Value: &value,
		}

		data, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Error marshal gauge %s: %v", name, err)
			return err
		}

		url := fmt.Sprintf("%s/update/", serverAddress)
		if err := sendRequest(client, url, data); err != nil {
			log.Printf("Error sending gauge %s: %v", name, err)
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
			log.Printf("Error marshal gauge %s: %v", name, err)
			return err
		}

		url := fmt.Sprintf("%s/update/", serverAddress)
		if err := sendRequest(client, url, data); err != nil {
			log.Printf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	return nil
}

func SendAll(storage storage.StorageIface, serverAddress string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s/updates/", serverAddress)
	allMetrics := make([]*models.MetricJSON, 0)
	batch := make([]*models.MetricJSON, 0, 100)
	metrics, err := storage.GetMetrics(context.Background())
	if err != nil {
		return err
	}
	for key, v := range metrics.Gauges {
		m, err := utils.MarshalMetricToJSON(models.Gauge, key, v)
		if err == nil {
			allMetrics = append(allMetrics, m)
		}
	}

	for key, v := range metrics.Counters {
		m, err := utils.MarshalMetricToJSON(models.Counter, key, v)
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

			err = sendRequest(client, url, data)
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
		cbody, err := utils.CompressGzip(body)
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
	return utils.WithRetry(context.Background(), operation)
}
