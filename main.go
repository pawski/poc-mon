package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pawski/poc-mon/internal/configuration"
	"github.com/pawski/poc-mon/internal/telemetry"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	appConfig, err := configuration.GetApp()
	if err != nil {
		logger.Fatal(err)
	}

	envConfig, err := configuration.GetEnv()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("Starting")

	reg := prometheus.NewRegistry()

	telemetry.Observations = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "observer",
		Subsystem:   "scrapes",
		Name:        "total",
		Help:        "Total number of scrapes occurred",
		ConstLabels: map[string]string{"host": "localhost"},
	})
	reg.MustRegister(telemetry.Observations)

	telemetry.ObservationsBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "observer",
			Subsystem: "scrapes",
			Name:      "bytes_total",
			Help:      "Total number of bytes returned by scrape",
		},
		[]string{"target"},
	)
	reg.MustRegister(telemetry.ObservationsBytes)

	telemetry.Duration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "observer",
			Subsystem: "test",
			Name:      "http_call_duration",
			Help:      "Time of http test call",
		},
		[]string{"host", "target", "code"},
	)
	reg.MustRegister(telemetry.Duration)

	if appConfig.EnableInternalMetrics {
		reg.MustRegister(collectors.NewBuildInfoCollector())
		reg.MustRegister(collectors.NewGoCollector(collectors.WithGoCollections(
			collectors.GoRuntimeMemStatsCollection | collectors.GoRuntimeMetricsCollection,
		)))
	}

	h := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(writer http.ResponseWriter, request *http.Request) {
		telemetry.Observations.Inc()
		h.ServeHTTP(writer, request)
	})

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("Bazinga!!!"))
		if err != nil {
			logger.Printf("%s", err)
		}
	})

	srv := &http.Server{
		Addr:    envConfig.HttpServerAddress,
		Handler: mux,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	stopTest := make(chan struct{}, 1)
	go func() {
		defer func() {
			wg.Done()
		}()
		client := http.Client{}
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-stopTest:
				return
			case <-ticker.C:
				request, err := http.NewRequest("GET", appConfig.TestUrl, nil) // HEAD maybe?
				if err != nil {
					logger.Printf("%s", err)
					continue
				}

				start := time.Now()
				var testStatus string
				response, err := client.Do(request)
				if err != nil {
					logger.Printf("%s", err)
					testStatus = ""
				}
				testStatus = response.Status
				telemetry.Duration.WithLabelValues("localhost", appConfig.TestUrl, testStatus).Set(time.Since(start).Seconds())
				b, err := io.ReadAll(response.Body)
				if err != nil {
					logger.Printf("%s", err)
					continue
				}
				telemetry.ObservationsBytes.WithLabelValues(appConfig.TestUrl).Add(float64(len(b)))
			}
		}
	}()

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			logger.Printf("%s", err)
		}
	}()

	logger.Printf("Running on %s...", envConfig.HttpServerAddress)
	<-shutdown
	stopTest <- struct{}{}
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown Failed:%+v", err)
	}

	logger.Printf("Shutting down...")
}
