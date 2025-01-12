package main

import (
	"flag"
	"log"

	"github.com/runtime-metrics-course/internal/server"
	"github.com/runtime-metrics-course/internal/storage"
)

func main() {
	address := flag.String("a", "localhost:8080", "server address ")
	flag.Parse()
	if err := storage.InitStorage(storage.RuntimeMemory); err != nil {
		log.Fatal(err)
	}
	if err := server.InitSever(*address); err != nil {
		log.Fatal(err)
	}

}
