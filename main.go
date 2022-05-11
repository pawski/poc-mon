package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("Starting")

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector())

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writtenBytes, err := writer.Write([]byte("Bazinga!!!"))
		if err != nil {
			logger.Printf("%s", err)
		}
		logger.Printf("%s", writtenBytes)
	})

	go func() {
		err := http.ListenAndServe("localhost:8000", mux)
		if err != http.ErrServerClosed {
			fmt.Println()
		}

	}()

	logger.Printf("Running...")
	<-shutdown
	logger.Printf("Shutting down...")
}
