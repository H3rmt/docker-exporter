package main

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"docker-exporter/internal/exporter"
	"docker-exporter/internal/status"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Docker Prometheus exporter...")

	// Initialize Docker client and metrics
	dockerClient, err := exporter.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	var reg prometheus.Gatherer
	if os.Getenv("ENABLE_INTERNAL_METRICS") == "true" {
		reg = prometheus.DefaultGatherer
		exporter.RegisterCollectorsWithRegistry(dockerClient, nil)
	} else {
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		registry := prometheus.NewRegistry()
		exporter.RegisterCollectorsWithRegistry(dockerClient, registry)
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		reg = registry
	}
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.Handle("/status", status.HandleStatus(dockerClient))

	server := &http.Server{Addr: ":9100"}

	go func() {
		log.Println("Listening on :9100/metrics")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down exporter...")
	err = server.Close()
	if err != nil {
		log.Fatalf("Failed to close HTTP server: %v", err)
	}
}
