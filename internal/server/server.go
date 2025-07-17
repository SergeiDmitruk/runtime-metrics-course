package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/middleware"
	"github.com/runtime-metrics-course/internal/storage"
)

// InitServer initializes and starts the HTTP server with configured routes and middleware.
//
// Parameters:
//   - address: Server listen address (e.g. ":8080")
//   - secretKey: Secret key for request authentication (empty disables auth)
//
// Returns:
//   - error if server fails to start
//
// Routes configured:
//   - GET / - Main metrics endpoint
//   - GET /ping - Database health check
//   - POST /updates/ - Batch update metrics
//   - /value/ - Metric retrieval endpoints
//   - /update/ - Metric update endpoints
//
// Middleware applied:
//   - Request logging
//   - Response compression
//   - HMAC authentication (if secretKey provided)
func InitServer(address, secretKey, cryptoKeyPath string) error {
	storage, err := storage.GetStorageManager().GetStorage()
	if err != nil {
		return err
	}

	r := chi.NewRouter()

	// Apply middleware stack
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.CompressMiddleware)
	if secretKey != "" {
		r.Use(middleware.NewHashMiddleware([]byte(secretKey)).Middleware)
	}
	if cryptoKeyPath != "" {
		cryptoMiddleware, err := middleware.NewCryptoMiddleware(cryptoKeyPath)
		if err != nil {
			return fmt.Errorf("failed to init crypto middleware: %w", err)
		}
		r.Use(cryptoMiddleware.Middleware)
	}
	// Initialize metrics handler
	mh := NewMetricsHandler(storage)

	// Configure routes
	r.Mount("/debug", pprofRouter())
	r.Get("/", mh.GetMetrics)
	r.Get("/ping", mh.PingDBHandler)
	r.Post("/updates/", mh.UpdateAll)

	// Metric value routes
	r.Route("/value/", func(r chi.Router) {
		r.Post("/", mh.GetMetricValueJSON)                // JSON endpoint
		r.Get("/{metric_type}/{name}", mh.GetMetricValue) // Plaintext endpoint
	})

	// Metric update routes
	r.Route("/update/", func(r chi.Router) {
		r.Post("/", mh.UpdateJSON)                         // JSON endpoint
		r.Post("/{metric_type}/{name}/{value}", mh.Update) // Plaintext endpoint
	})

	logger.Log.Sugar().Infoln("Server starting on", address)
	return http.ListenAndServe(address, r)
}

func pprofRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/pprof/*", http.HandlerFunc(pprof.Index))
	r.Get("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Get("/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Get("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Get("/pprof/trace", http.HandlerFunc(pprof.Trace))
	return r
}
