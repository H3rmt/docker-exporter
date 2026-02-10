package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/exporter"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/h3rmt/docker-exporter/internal/status"
	"github.com/h3rmt/docker-exporter/internal/web"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	verbose           bool
	trace             bool
	quiet             bool
	internalMetrics   bool
	logFormat         string
	homepage          bool
	sizeCacheDuration time.Duration
	address           string
	port              string
	dockerHost        string
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "Docker Prometheus exporter",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Enum check
		switch logFormat {
		case "json", "logfmt":
		default:
			return fmt.Errorf("invalid --log-format: %s (want json|logfmt)", logFormat)
		}
		return nil
	},
	Run: run,
}

func init() {
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode (enabled debug logs).")
	rootCmd.Flags().BoolVar(&trace, "trace", false, "Very Verbose mode (enabled trace logs).")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (disables info logs).")
	rootCmd.Flags().BoolVar(&internalMetrics, "internal-metrics", false, "Enable internal go metrics.")
	rootCmd.Flags().StringVar(&logFormat, "log-format", "logfmt", "Log format: 'logfmt' or 'json'.")
	rootCmd.Flags().BoolVar(&homepage, "homepage", true, "Show homepage with charts.")
	rootCmd.Flags().DurationVar(&sizeCacheDuration, "size-cache-duration", time.Duration(300)*time.Second, "Duration to wait before refreshing container size cache.")
	rootCmd.Flags().StringVarP(&address, "address", "a", "0.0.0.0", "Address to listen on.")
	rootCmd.Flags().StringVarP(&port, "port", "p", "9100", "Port to listen on.")
	rootCmd.Flags().StringVarP(&dockerHost, "docker-host", "d", "unix:///var/run/docker.sock", "Host to connect to.")
}

var (
	// Version can be set at build time using -ldflags "-X main.Version=vx.y.z"
	Version = "main"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) {
	log.InitLogger(logFormat, verbose, trace, quiet)
	log.GetLogger().Info("Starting Docker Prometheus exporter",
		"version", Version,
		"uid", os.Getuid(),
		"gid", os.Getgid(),
		"docker_host", dockerHost,
		"log_format", logFormat,
	)
	docker.SetSizeCacheDuration(sizeCacheDuration)

	if os.Getenv("IP") == "" {
		log.GetLogger().Info("IP environment variable not set, pass it do display the IP of the exporter on the homepage", "missing_env", "IP")
	}

	// Initialize Docker client and metrics
	dockerClient, err := docker.NewDockerClient(dockerHost)
	if err != nil {
		log.GetLogger().Error("Failed to create Docker client", "error", err, "docker_host", dockerHost)
		os.Exit(1)
	}

	log.GetLogger().Info("Initializing Docker Prometheus exporter...")
	var reg prometheus.Gatherer
	if internalMetrics {
		reg = prometheus.DefaultGatherer
		exporter.RegisterCollectorsWithRegistry(dockerClient, nil, Version)
	} else {
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		registry := prometheus.NewRegistry()
		exporter.RegisterCollectorsWithRegistry(dockerClient, registry, Version)
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		reg = registry
	}

	// Wrapper for /metrics that returns 503 when not ready
	metricsHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !status.IsReady() {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("503 Service Unavailable - Collecting initial metrics, please wait...\n"))
			return
		}
		metricsHandler.ServeHTTP(w, r)
	}))
	http.Handle("/status", status.HandleStatus(dockerClient, Version))
	// Web UI and API
	if homepage {
		http.HandleFunc("/", web.HandleRoot())
		http.HandleFunc("/api/info", web.HandleAPIInfo(Version))
		http.HandleFunc("/api/usage", web.HandleAPIUsage())
		http.Handle("/api/containers", web.HandleAPIContainers(dockerClient))

		go func() {
			web.CollectInBg()
			log.GetLogger().Debug("Metrics in background collector stopped")
		}()
	} else {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("404 page not found (homepage disabled)\n"))
		})
	}

	server := &http.Server{Addr: fmt.Sprintf("%s:%s", address, port), ErrorLog: slog.NewLogLogger(log.GetLogger().Handler(), slog.LevelWarn)}
	log.GetLogger().Info("HTTP server created")
	go func() {
		log.GetLogger().Info("Listening on metrics endpoint", "address", fmt.Sprintf("%s:%s", address, port))
		if err := server.ListenAndServe(); err != nil {
			log.GetLogger().Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Collect initial metrics in background
	go func() {
		log.GetLogger().Info("Collecting initial metrics in background...")
		ctx := context.Background()
		
		// Perform an initial collection to warm up caches
		containers, err := dockerClient.ListAllRunningContainers(ctx)
		if err != nil {
			log.GetLogger().Warn("Initial container listing failed", "error", err)
		} else {
			// Warm up by inspecting and getting stats for all containers
			log.GetLogger().Debug("Warming up container stats cache", "count", len(containers))
			for _, container := range containers {
				_, _ = dockerClient.InspectContainer(ctx, container.ID, true)
				_, _ = dockerClient.GetContainerStats(ctx, container.ID)
			}
		}
		
		log.GetLogger().Info("Initial metrics collection complete")
		status.SetReady()
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
