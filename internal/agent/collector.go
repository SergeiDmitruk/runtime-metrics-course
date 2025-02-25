package agent

import (
	"context"
	"math/rand"
	"runtime"

	"github.com/runtime-metrics-course/internal/storage"
)

func CollectRuntimeMetrics(storage storage.StorageIface) {
	ctx := context.Background()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	storage.UpdateGauge(ctx, "Alloc", float64(memStats.Alloc))
	storage.UpdateGauge(ctx, "BuckHashSys", float64(memStats.BuckHashSys))
	storage.UpdateGauge(ctx, "Frees", float64(memStats.Frees))
	storage.UpdateGauge(ctx, "GCCPUFraction", memStats.GCCPUFraction)
	storage.UpdateGauge(ctx, "GCSys", float64(memStats.GCSys))
	storage.UpdateGauge(ctx, "HeapAlloc", float64(memStats.HeapAlloc))
	storage.UpdateGauge(ctx, "HeapIdle", float64(memStats.HeapIdle))
	storage.UpdateGauge(ctx, "HeapInuse", float64(memStats.HeapInuse))
	storage.UpdateGauge(ctx, "HeapObjects", float64(memStats.HeapObjects))
	storage.UpdateGauge(ctx, "HeapReleased", float64(memStats.HeapReleased))
	storage.UpdateGauge(ctx, "HeapSys", float64(memStats.HeapSys))
	storage.UpdateGauge(ctx, "LastGC", float64(memStats.LastGC))
	storage.UpdateGauge(ctx, "Lookups", float64(memStats.Lookups))
	storage.UpdateGauge(ctx, "MCacheInuse", float64(memStats.MCacheInuse))
	storage.UpdateGauge(ctx, "MCacheSys", float64(memStats.MCacheSys))
	storage.UpdateGauge(ctx, "MSpanInuse", float64(memStats.MSpanInuse))
	storage.UpdateGauge(ctx, "MSpanSys", float64(memStats.MSpanSys))
	storage.UpdateGauge(ctx, "Mallocs", float64(memStats.Mallocs))
	storage.UpdateGauge(ctx, "NextGC", float64(memStats.NextGC))
	storage.UpdateGauge(ctx, "NumForcedGC", float64(memStats.NumForcedGC))
	storage.UpdateGauge(ctx, "NumGC", float64(memStats.NumGC))
	storage.UpdateGauge(ctx, "OtherSys", float64(memStats.OtherSys))
	storage.UpdateGauge(ctx, "PauseTotalNs", float64(memStats.PauseTotalNs))
	storage.UpdateGauge(ctx, "StackInuse", float64(memStats.StackInuse))
	storage.UpdateGauge(ctx, "StackSys", float64(memStats.StackSys))
	storage.UpdateGauge(ctx, "Sys", float64(memStats.Sys))
	storage.UpdateGauge(ctx, "TotalAlloc", float64(memStats.TotalAlloc))
	storage.UpdateGauge(ctx, "RandomValue", rand.Float64())

	storage.UpdateCounter(ctx, "PollCount", 1)
}
