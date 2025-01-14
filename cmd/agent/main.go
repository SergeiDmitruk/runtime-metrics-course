package main

import (
	"errors"
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/runtime-metrics-course/internal/agent"
	"github.com/runtime-metrics-course/internal/storage"
)

type Address struct {
	Host string
	Port int
}

func (a Address) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}
func (a *Address) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("Need address in a form host:port")
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
	addr := &Address{
		Host: "localhost",
		Port: 8080,
	}
	_ = flag.Value(addr)
	//address := flag.String("a", "http://localhost:8080", "server address ")
	flag.Var(addr, "a", "Net address host:port")
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
	if err := agent.StartAgent(storage, addr.String(), time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Println("agent start on", addr.String())
	select {}
}
