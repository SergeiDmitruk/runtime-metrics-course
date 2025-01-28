package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/middleware"
	"github.com/runtime-metrics-course/internal/storage"
)

func InitSever(address string) error {

	storage, err := storage.GetStorage()
	if err != nil {
		return err
	}
	r := chi.NewRouter()
	r.Get("/", logger.LoggerMdlwr(middleware.CompressMdlwr(GetMetricsHandler(storage))))
	r.Route("/value/", func(r chi.Router) {
		r.Post("/", logger.LoggerMdlwr(middleware.CompressMdlwr(GetMetricValueJSONHandler(storage))))
		r.Get("/{metric_type}/{name}", logger.LoggerMdlwr(middleware.CompressMdlwr(GetMetricValueHandler(storage))))
	})
	r.Route("/update/", func(r chi.Router) {
		r.Post("/", logger.LoggerMdlwr(middleware.CompressMdlwr(UpdateJSONHandler(storage))))
		r.Post("/{metric_type}/{name}/{value}", logger.LoggerMdlwr(middleware.CompressMdlwr(UpdateHandler(storage))))
	})

	log.Println("Server start on", address)
	return http.ListenAndServe(address, r)
}
