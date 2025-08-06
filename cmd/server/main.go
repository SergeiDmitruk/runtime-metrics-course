package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/server"
	"github.com/runtime-metrics-course/internal/storage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

type ServerConfig struct {
	Address       string        `json:"address"`
	SecretKey     string        `json:"key"`
	CryptoKey     string        `json:"crypto_key"`
	StoreInterval time.Duration `json:"store_interval"`
	FilePath      string        `json:"store_file"`
	Restore       bool          `json:"restore"`
	DatabaseDSN   string        `json:"database_dsn"`
	TrustedSubnet string        `json:"trusted_subnet"`
}

func printBuildInfo() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}

func main() {
	printBuildInfo()

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.Init("info"); err != nil {
		log.Fatal(err)
	}

	var conn *sql.DB
	if cfg.DatabaseDSN != "" {
		conn, err = initDB(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}
		defer conn.Close()
	}

	storageCfg := &storage.Cfg{
		Interval: cfg.StoreInterval,
		FilePath: cfg.FilePath,
		Restore:  cfg.Restore,
		Conn:     conn,
	}

	sm, err := storage.NewStorageManager(storageCfg)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	sm.SaverRun()

	if err := server.InitServer(cfg.Address, cfg.SecretKey, cfg.CryptoKey, cfg.TrustedSubnet); err != nil {
		logger.Log.Fatal(err.Error())
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-sigChan
	fmt.Println("Завершение работы...")
	sm.SaverStop()
}

func LoadConfig() (*ServerConfig, error) {

	cfg := &ServerConfig{
		Address:       "localhost:8080",
		StoreInterval: 300 * time.Second,
		FilePath:      "metrics.json",
		Restore:       true,
	}

	var configFile string

	if configFile == "" {
		configFile = os.Getenv("CONFIG")
	}

	if configFile != "" {
		fileData, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var fileCfg ServerConfig
		if err := json.Unmarshal(fileData, &fileCfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		if fileCfg.Address != "" {
			cfg.Address = fileCfg.Address
		}
		if fileCfg.StoreInterval != 0 {
			cfg.StoreInterval = fileCfg.StoreInterval
		}
		if fileCfg.FilePath != "" {
			cfg.FilePath = fileCfg.FilePath
		}
		if fileCfg.DatabaseDSN != "" {
			cfg.DatabaseDSN = fileCfg.DatabaseDSN
		}
		if fileCfg.CryptoKey != "" {
			cfg.CryptoKey = fileCfg.CryptoKey
		}
		if fileCfg.SecretKey != "" {
			cfg.SecretKey = fileCfg.SecretKey
		}
		if fileCfg.TrustedSubnet != "" {
			cfg.TrustedSubnet = fileCfg.TrustedSubnet
		}
		cfg.Restore = fileCfg.Restore
	}

	flag.StringVar(&configFile, "c", "", "Path to config file")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.StringVar(&cfg.Address, "a", cfg.Address, "server address")
	flag.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "ключ шифрования")
	flag.DurationVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "Интервал сохранения в секундах (0 = синхронное сохранение)")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "путь к файлу с приватным ключом")
	flag.StringVar(&cfg.FilePath, "f", cfg.FilePath, "Путь до файла хранения метрик")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "Восстанавливать метрики при старте")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DB DSN")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Подсеть для доступа к API")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Address = envAddr
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.SecretKey = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}
	if envStoreInt := os.Getenv("STORE_INTERVAL"); envStoreInt != "" {
		if val, err := strconv.Atoi(envStoreInt); err == nil {
			cfg.StoreInterval = time.Duration(val) * time.Second
		}
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FilePath = envFilePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if val, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = val
		}
	}
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DatabaseDSN = envDSN
	}
	if envSubnet := os.Getenv("TRUSTED_SUBNET"); envSubnet != "" {
		cfg.TrustedSubnet = envSubnet
	}
	if cfg.TrustedSubnet != "" {
		cfg.TrustedSubnet = strings.TrimSpace(cfg.TrustedSubnet)

		_, _, err := net.ParseCIDR(strings.TrimSpace(cfg.TrustedSubnet))
		if err != nil {
			return nil, fmt.Errorf("invalid subnet: %s", cfg.TrustedSubnet)
		}

	}
	return cfg, nil
}

func initDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(conn, "internal/migrations"); err != nil {
		return nil, fmt.Errorf("ошибка применения миграций: %w", err)
	}

	logger.Log.Sugar().Info("Подключение к БД успешно")
	return conn, nil
}
