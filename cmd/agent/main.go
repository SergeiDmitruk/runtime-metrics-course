package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/storage"
)

type Config struct {
	Host           string
	Port           int
	PollInterval   int `env:"REPORT_INTERVAL"`
	ReportInterval int `env:"POLL_INTERVAL"`
}

func (a Config) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}
func (a *Config) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need Config in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}
func main() {
	config := &Config{
		Host: "localhost",
		Port: 8080,
	}
	_ = flag.Value(config)
	//Config := flag.String("a", "http://localhost:8080", "server Config ")
	flag.Var(config, "a", "Net Config host:port")
	config.PollInterval = *flag.Int("p", 2, "poll interval")
	config.ReportInterval = *flag.Int("r", 10, "report interval")

	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		log.Fatal(err)
	}

	if configEnv := os.Getenv("ADDRESS"); configEnv != "" {
		if err := config.Set(configEnv); err != nil {
			log.Fatal(err)
		}
	}

	if !strings.Contains(config.String(), "http") {
		config.Host = "http://" + config.Host
	}
	if err := storage.InitStorage(storage.RuntimeMemory); err != nil {
		log.Fatal(err)
	}
	storage, err := storage.GetStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := agent.StartAgent(storage, config.String(), time.Duration(config.PollInterval)*time.Second, time.Duration(config.ReportInterval)*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Println("agent start on", config.String())
	select {}
}
