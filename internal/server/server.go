package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/runtime-metrics-course/internal/storage"
)

func InitSever() error {

	storage, err := storage.GetStorage()
	if err != nil {
		return err
	}
	r := chi.NewRouter()
	r.Get("/", GetMetricsHandler(storage))
	r.Get("/value/{metric_type}/{name}", GetMetricValueHandler(storage))
	r.Route("/update/{metric_type}/", func(r chi.Router) {
		r.Post("/", http.NotFound)
		r.Post("/{name}/{value}", UpdateHandler(storage))
	})

	log.Println("Server start")
	return http.ListenAndServe(":8080", r)
}
