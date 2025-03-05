package agent

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/runtime-metrics-course/internal/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func CollectRuntimeMetrics(ch chan<- Task) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memStat := map[string]float64{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": memStats.GCCPUFraction,
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}

	for name, value := range memStat {
		ch <- Task{Metric: models.MetricJSON{ID: name, MType: models.Gauge, Value: &value}}
	}

}

func CollectGoupsutiMetrics(ch chan<- Task) {
	var tasks []Task
	v, _ := mem.VirtualMemory()
	total := float64(v.Total)
	free := float64(v.Free)
	cpuUtilization := 0.0
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		cpuUtilization = cpuPercent[0]
	}
	tasks = append(tasks, Task{Metric: models.MetricJSON{ID: "TotalMemory", MType: models.Gauge, Value: &total}},
		Task{Metric: models.MetricJSON{ID: "FreeMemory", MType: models.Gauge, Value: &free}},
		Task{Metric: models.MetricJSON{ID: "CPUutilization1", MType: models.Gauge, Value: &cpuUtilization}})
	for _, task := range tasks {
		ch <- task
	}
}
