package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/server"
	"github.com/runtime-metrics-course/internal/storage"
)

var address string
var secretKey string
var databaseDSN string
var conn *sql.DB

func main() {
	cfg := ParseFlags()

	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}
	if databaseDSN != "" {
		if err := initDB(databaseDSN); err != nil {
			logger.Log.Fatal(err.Error())
		}
	}
	defer conn.Close()
	cfg.Conn = conn
	sm, err := storage.NewStorageManager(cfg)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	sm.SaverRun()

	if err := server.InitServer(address, secretKey); err != nil {
		logger.Log.Fatal(err.Error())
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Завершение работы...")
	sm.SaverStop()
}

func ParseFlags() *storage.Cfg {
	addressFlag := flag.String("a", "localhost:8080", "server address ")
	keyFlag := flag.String("k", "", "ключ шифрования")
	storeInterval := flag.Int("i", 300, "Интервал сохранения в секундах (0 = синхронное сохранение)")
	filePath := flag.String("f", "metrics.json", "Путь до файла хранения метрик")
	restore := flag.Bool("r", true, "Восстанавливать метрики при старте")
	databaseDSNFlag := flag.String("d", "", "DB DSN")

	flag.Parse()
	address = *addressFlag
	secretKey = *keyFlag
	databaseDSN = *databaseDSNFlag
	if env, ok := os.LookupEnv("ADDRESS"); ok {
		address = env
	}
	if env, ok := os.LookupEnv("KEY"); ok {
		secretKey = env
	}
	if env, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if val, err := strconv.Atoi(env); err == nil {
			*storeInterval = val
		}
	}
	if env, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		*filePath = env
	}
	if env, ok := os.LookupEnv("RESTORE"); ok {
		if val, err := strconv.ParseBool(env); err == nil {
			*restore = val
		}
	}
	if env, ok := os.LookupEnv("DATABASE_DSN"); ok {
		databaseDSN = env
	}
	Cfg := &storage.Cfg{
		Interval: time.Duration(*storeInterval) * time.Second,
		FilePath: *filePath,
		Restore:  *restore,
	}

	return Cfg
}

func initDB(dsn string) error {
	var err error

	conn, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(conn, "internal/migrations"); err != nil {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	logger.Log.Sugar().Info("Подключение к БД успешно")
	return nil
}
