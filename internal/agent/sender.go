package agent

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/runtime-metrics-course/internal/storage"
)

func SendMetrics(storage storage.StorageIface, serverAddress string) error {
	log.Println("------Sending metrics------")

	client := &http.Client{Timeout: 5 * time.Second}
	metrics := storage.GetMetrics()

	for name, value := range metrics.Gauges {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", serverAddress, name, value)
		if err := sendRequest(client, url); err != nil {
			log.Printf("Error sending gauge %s: %v", name, err)
			return err
		}
	}

	for name, value := range metrics.Counters {
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverAddress, name, value)
		if err := sendRequest(client, url); err != nil {
			log.Printf("Error sending counter %s: %v", name, err)
			return err
		}
	}

	log.Println("------Metrics sent successfully------")
	return nil
}

func sendRequest(client *http.Client, url string) error {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
