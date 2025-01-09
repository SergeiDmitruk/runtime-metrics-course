package main

import (
	"log"

	"github.com/runtime-metrics-course/internal/server"
	"github.com/runtime-metrics-course/internal/storage"
)

func main() {
	if err := storage.InitStorage(storage.RuntimeMemory); err != nil {
		log.Fatal(err)
	}
	if err := server.InitSever(); err != nil {
		log.Fatal(err)
	}

}
