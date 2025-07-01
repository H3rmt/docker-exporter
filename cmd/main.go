package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"docker-exporter/internal/exporter"
	"docker-exporter/internal/log"
	"docker-exporter/internal/status"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	verbose         = kingpin.Flag("verbose", "Verbose mode (enabled debug logs).").Short('v').Default("false").Bool()
	quiet           = kingpin.Flag("quiet", "Quiet mode (disables info logs).").Short('q').Default("false").Bool()
	internalMetrics = kingpin.Flag("internal-metrics", "Enable internal metrics.").Default("false").Bool()
	address         = kingpin.Flag("address", "Address to listen on.").Short('a').Default("0.0.0.0").String()
	port            = kingpin.Flag("port", "Port to listen on.").Short('p').Default("9100").String()
)

func main() {
	kingpin.Parse()
	log.SetVerbose(*verbose)
	log.SetQuiet(*quiet)

	log.Info("Starting Docker Prometheus exporter...")

	// Initialize Docker client and metrics
	dockerClient, err := exporter.NewDockerClient()
	if err != nil {
		log.Error("Failed to create Docker client: %v", err)
	}

	var reg prometheus.Gatherer
	if *internalMetrics {
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

	server := &http.Server{Addr: fmt.Sprintf("%s:%s", *address, *port), ErrorLog: log.WarningLogger}

	go func() {
		log.Info("Listening on :9100/metrics")
		if err := server.ListenAndServe(); err != nil {
			log.Error("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info("Shutting down exporter...")
	err = server.Close()
	if err != nil {
		log.Error("Failed to close HTTP server: %v", err)
	}
}
