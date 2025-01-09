package agent

import (
	"math/rand"
	"runtime"

	"github.com/runtime-metrics-course/internal/storage"
)

func CollectRuntimeMetrics(storage storage.StorageIface) {

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	storage.UpdateGauge("Alloc", float64(memStats.Alloc))
	storage.UpdateGauge("BuckHashSys", float64(memStats.BuckHashSys))
	storage.UpdateGauge("Frees", float64(memStats.Frees))
	storage.UpdateGauge("GCCPUFraction", memStats.GCCPUFraction)
	storage.UpdateGauge("GCSys", float64(memStats.GCSys))
	storage.UpdateGauge("HeapAlloc", float64(memStats.HeapAlloc))
	storage.UpdateGauge("HeapIdle", float64(memStats.HeapIdle))
	storage.UpdateGauge("HeapInuse", float64(memStats.HeapInuse))
	storage.UpdateGauge("HeapObjects", float64(memStats.HeapObjects))
	storage.UpdateGauge("HeapReleased", float64(memStats.HeapReleased))
	storage.UpdateGauge("HeapSys", float64(memStats.HeapSys))
	storage.UpdateGauge("LastGC", float64(memStats.LastGC))
	storage.UpdateGauge("Lookups", float64(memStats.Lookups))
	storage.UpdateGauge("MCacheInuse", float64(memStats.MCacheInuse))
	storage.UpdateGauge("MCacheSys", float64(memStats.MCacheSys))
	storage.UpdateGauge("MSpanInuse", float64(memStats.MSpanInuse))
	storage.UpdateGauge("MSpanSys", float64(memStats.MSpanSys))
	storage.UpdateGauge("Mallocs", float64(memStats.Mallocs))
	storage.UpdateGauge("NextGC", float64(memStats.NextGC))
	storage.UpdateGauge("NumForcedGC", float64(memStats.NumForcedGC))
	storage.UpdateGauge("NumGC", float64(memStats.NumGC))
	storage.UpdateGauge("OtherSys", float64(memStats.OtherSys))
	storage.UpdateGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	storage.UpdateGauge("StackInuse", float64(memStats.StackInuse))
	storage.UpdateGauge("StackSys", float64(memStats.StackSys))
	storage.UpdateGauge("Sys", float64(memStats.Sys))
	storage.UpdateGauge("TotalAlloc", float64(memStats.TotalAlloc))
	storage.UpdateGauge("RandomValue", rand.Float64())

	storage.UpdateCounter("PollCount", 1)
	//	log.Println(storage.GetGauges())
}
