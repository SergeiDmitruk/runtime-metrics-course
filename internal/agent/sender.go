package agent

import (
	"bytes"
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
	metrics := storage.GetMetrics()

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
	metrics := storage.GetMetrics()

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

func sendRequest(client *http.Client, url string, body []byte) error {
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
