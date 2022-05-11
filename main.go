package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	srv := &http.Server{
		Addr:    "localhost:8000",
		Handler: mux,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			fmt.Println()
		}
	}()

	logger.Printf("Running...")
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown Failed:%+v", err)
	}

	logger.Printf("Shutting down...")
}
