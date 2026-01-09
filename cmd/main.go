package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/exporter"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/h3rmt/docker-exporter/internal/status"
	"github.com/h3rmt/docker-exporter/internal/web"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Version can be set at build time using -ldflags "-X main.Version=x.y.z"
	Version = "dev"

	verbose          = kingpin.Flag("verbose", "Verbose mode (enabled debug logs).").Short('v').Default("false").Bool()
	trace            = kingpin.Flag("trace", "Very Verbose mode (enabled trace logs).").Default("false").Bool()
	quiet            = kingpin.Flag("quiet", "Quiet mode (disables info logs).").Short('q').Default("false").Bool()
	logFormat        = kingpin.Flag("log-format", "Log format: 'logfmt' or 'json'.").Default("logfmt").Enum("logfmt", "json")
	internalMetrics  = kingpin.Flag("internal-metrics", "Enable internal go metrics.").Default("false").Bool()
	address          = kingpin.Flag("address", "Address to listen on.").Short('a').Default("0.0.0.0").String()
	secondsCacheSize = kingpin.Flag("size-cache-seconds", "Seconds to wait before refreshing container size cache.").Default("300").Int()
	port             = kingpin.Flag("port", "Port to listen on.").Short('p').Default("9100").String()
	dockerHost       = kingpin.Flag("docker-host", "Host to connect to.").Short('d').Default("unix:///var/run/docker.sock").String()
)

func main() {
	kingpin.Parse()
	log.InitLogger(*logFormat, *verbose, *trace, *quiet)
	log.GetLogger().Info("Starting Docker Prometheus exporter",
		"version", Version,
		"uid", os.Getuid(),
		"gid", os.Getgid(),
		"docker_host", *dockerHost,
		"log_format", *logFormat,
	)
	docker.SetSizeCacheSeconds(time.Duration(*secondsCacheSize) * time.Second)

	// Initialize Docker client and metrics
	dockerClient, err := docker.NewDockerClient(*dockerHost)
	if err != nil {
		log.GetLogger().Error("Failed to create Docker client", "error", err, "docker_host", *dockerHost)
		os.Exit(1)
	}

	log.GetLogger().Info("Collecting initial metrics...")
	var reg prometheus.Gatherer
	if *internalMetrics {
		reg = prometheus.DefaultGatherer
		exporter.RegisterCollectorsWithRegistry(dockerClient, nil, Version)
	} else {
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		registry := prometheus.NewRegistry()
		exporter.RegisterCollectorsWithRegistry(dockerClient, registry, Version)
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		reg = registry
	}

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.Handle("/status", status.HandleStatus(dockerClient))
	// Web UI and API
	http.HandleFunc("/", web.HandleRoot())
	http.HandleFunc("/api/info", web.HandleAPIInfo(Version))
	http.HandleFunc("/api/usage", web.HandleAPIUsage())
	http.Handle("/api/containers", web.HandleAPIContainers(dockerClient))

	go func() {
		web.CollectInBg()
		log.GetLogger().Debug("Metrics in background collector stopped")
	}()

	server := &http.Server{Addr: fmt.Sprintf("%s:%s", *address, *port), ErrorLog: slog.NewLogLogger(log.GetLogger().Handler(), slog.LevelWarn)}
	log.GetLogger().Info("HTTP server created")
	go func() {
		log.GetLogger().Info("Listening on metrics endpoint", "address", fmt.Sprintf("%s:%s", *address, *port))
		if err := server.ListenAndServe(); err != nil {
			log.GetLogger().Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	web.StopCollect()
	log.GetLogger().Info("Shutting down exporter...")
	err = server.Close()
	if err != nil {
		log.GetLogger().Error("Failed to close HTTP server", "error", err)
		os.Exit(1)
	}
}
