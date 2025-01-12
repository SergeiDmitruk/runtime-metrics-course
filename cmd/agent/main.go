package main

import (
	"flag"
	"log"
	"time"

	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/storage"
)

func main() {
	address := flag.String("a", "http://localhost:8080", "server address ")
	pollInterval := flag.Int("p", 2, "poll interval")
	reportInterval := flag.Int("r", 10, "report interval")

	flag.Parse()
	if err := storage.InitStorage(storage.RuntimeMemory); err != nil {
		log.Fatal(err)
	}
	storage, err := storage.GetStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := agent.StartAgent(storage, *address, time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Println("agent start")
	select {}
}
