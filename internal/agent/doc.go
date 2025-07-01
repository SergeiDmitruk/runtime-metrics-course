// Package agent implements metrics collection and reporting functionality.
//
// Key Features:
//
// * Collection of Go runtime metrics (memory, GC, goroutines)
// * Collection of system metrics (CPU, memory) via gopsutil
// * Multiple reporting formats:
//   - URL-encoded (text)
//   - JSON
//   - Batch updates
//
// * Flexible configuration:
//   - Collection/reporting intervals
//   - Rate limiting
//   - Request signing (HMAC)
//
// * Reliable delivery:
//   - Gzip compression
//   - Retry mechanism
//   - Worker pools
//
// Usage Example:
//
//	config := agent.Config{
//	    Host:           "localhost:8080",
//	    PollInterval:   2 * time.Second,
//	    ReportInterval: 10 * time.Second,
//	    RateLimit:      5,
//	}
//
//	if err := agent.StartAgent(config); err != nil {
//	    log.Fatal(err)
//	}
//
// Core Components:
//
// * Collector - metrics collection (runtime and system)
// * Sender - metrics reporting
// * WorkerPool - concurrent request handling
// * RetryMechanism - failed request retries
//
// Collected Metrics:
//
// Runtime metrics include:
// * Alloc - current memory allocations
// * HeapInuse - active heap memory
// * NumGC - garbage collection cycles
// * And others (see runtime.MemStats)
//
// System metrics include:
// * CPUutilization - CPU usage
// * TotalMemory - total system memory
// * FreeMemory - available memory
package agent
