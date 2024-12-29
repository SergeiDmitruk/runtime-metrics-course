package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MemStorage struct {
	mu       sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}
func (m *MemStorage) PrintMetrics() { // test stdout
	fmt.Println("------Metrics-------")
	for name, val := range m.gauges {
		fmt.Println(name, val)
	}
	for name, val := range m.counters {
		fmt.Println(name, val)
	}
	fmt.Println("--------------------")
}

func (m *MemStorage) handleUpdate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//path := strings.TrimPrefix(r.URL.Path, "/update/")
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 3 {
		if len(parts) == 2 && parts[1] == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	metricType, name, value := parts[0], parts[1], parts[2]

	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		m.UpdateGauge(name, val)
	case Counter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		m.UpdateCounter(name, val)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
	m.PrintMetrics()

	w.WriteHeader(http.StatusOK)
}

func main() {
	storage := NewMemStorage()
	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(storage.handleUpdate)))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
