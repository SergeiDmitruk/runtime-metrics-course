package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/middleware"
	"github.com/runtime-metrics-course/internal/storage"
)

func InitServer(address string) error {

	storage, err := storage.GetStorageManager().GetStorage()
	if err != nil {
		return err
	}

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.CompressMiddleware)
	mh := GetNewMetricsHandler(storage)
	r.Get("/", mh.GetMetrics)
	r.Get("/ping", mh.PingDBHandler)
	r.Post("/updates/", mh.UpdateAll)
	r.Route("/value/", func(r chi.Router) {
		r.Post("/", mh.GetMetricValueJSON)
		r.Get("/{metric_type}/{name}", mh.GetMetricValue)
	})
	r.Route("/update/", func(r chi.Router) {
		r.Post("/", mh.UpdateJSON)
		r.Post("/{metric_type}/{name}/{value}", mh.Update)
	})

	logger.Log.Sugar().Infoln("Server start on", address)
	return http.ListenAndServe(address, r)
}
