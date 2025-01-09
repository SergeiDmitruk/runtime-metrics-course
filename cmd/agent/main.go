package main

import (
	"log"
	"time"

	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/storage"
)

func main() {
	if err := storage.InitStorage(storage.RuntimeMemory); err != nil {
		log.Fatal(err)
	}
	storage, err := storage.GetStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := agent.StartAgent(storage, "http://localhost:8080", 2*time.Second, 10*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Println("agent start")
	select {}
}
