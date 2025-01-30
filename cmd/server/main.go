package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/server"
	"github.com/runtime-metrics-course/internal/storage"
)

func main() {
	address := flag.String("a", "localhost:8080", "server address ")
	storeInterval := flag.Int("i", 300, "Интервал сохранения в секундах (0 = синхронное сохранение)")
	filePath := flag.String("f", "metrics.json", "Путь до файла хранения метрик")
	restore := flag.Bool("r", true, "Восстанавливать метрики при старте")

	flag.Parse()

	if addr, ok := os.LookupEnv("ADDRESS"); ok {
		address = &addr
	}
	if env, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		fmt.Sscanf(env, "%d", storeInterval)
	}
	if env, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		*filePath = env
	}
	if env, ok := os.LookupEnv("RESTORE"); ok {
		fmt.Sscanf(env, "%t", restore)
	}

	sm, err := storage.NewStorageManager(storage.RuntimeMemory, time.Duration(*storeInterval)*time.Second, *filePath, *restore)
	if err != nil {
		log.Fatal(err)
	}
	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}
	sm.SaverRun()
	if err := server.InitSever(*address); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Завершение работы...")
	sm.SaverStop()

}
