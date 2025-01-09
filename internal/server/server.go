package server

import (
	"log"
	"net/http"

	"github.com/runtime-metrics-course/internal/storage"
)

func InitSever() error {
	mux := http.NewServeMux()
	storage, err := storage.GetStorage()
	if err != nil {
		return err
	}
	mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(UpdateHandler(storage))))
	log.Println("Server start")
	return http.ListenAndServe(":8080", mux)
}
