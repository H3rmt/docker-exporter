package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/h3rmt/docker-exporter/internal/glob"
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
	verbose                   bool
	trace                     bool
	quiet                     bool
	internalMetrics           bool
	logFormat                 string
	homepage                  bool
	sizeCacheDuration         time.Duration
	diskUsageCacheDuration    time.Duration
	address                   string
	port                      string
	dockerHost                string
	collectorSystem           bool
	collectorContainer        bool
	collectorContainerNetwork bool
	collectorContainerCPU     bool
	collectorContainerFS      bool
	collectorContainerStats   bool
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
	rootCmd.Flags().StringVarP(&dockerHost, "docker-host", "d", "unix:///var/run/docker.sock", "Host to connect to.")
	rootCmd.Flags().BoolVarP(&verbose, "log.verbose", "v", false, "Verbose mode (enabled debug logs).")
	rootCmd.Flags().BoolVar(&trace, "log.trace", false, "Very Verbose mode (enabled trace logs).")
	rootCmd.Flags().BoolVarP(&quiet, "log.quiet", "q", false, "Quiet mode (disables info logs).")
	rootCmd.Flags().StringVar(&logFormat, "log.format", "logfmt", "Log format: 'logfmt' or 'json'.")
	rootCmd.Flags().DurationVar(&sizeCacheDuration, "cache.size-cache-duration", time.Duration(300)*time.Second, "Duration to wait before refreshing container size cache.")
	rootCmd.Flags().DurationVar(&diskUsageCacheDuration, "cache.disk-usage-cache-seconds", time.Duration(120)*time.Second, "Duration to wait before refreshing docker disk usage cache.")
	rootCmd.Flags().BoolVar(&homepage, "web.homepage", true, "Show homepage with charts.")
	rootCmd.Flags().StringVarP(&address, "web.address", "a", "0.0.0.0", "Address to listen on.")
	rootCmd.Flags().StringVarP(&port, "web.port", "p", "9100", "Port to listen on.")
	rootCmd.Flags().BoolVar(&internalMetrics, "collector.internal-metrics", false, "Enable internal go metrics.")
	rootCmd.Flags().BoolVar(&collectorSystem, "collector.system", true, "Enable system collector (exporter info, host OS info).")
	rootCmd.Flags().BoolVar(&collectorContainer, "collector.container", true, "Enable container collector.")
	rootCmd.Flags().BoolVar(&collectorContainerNetwork, "collector.container.net", true, "Enable container network collector.")
	rootCmd.Flags().BoolVar(&collectorContainerCPU, "collector.container.cpu", true, "Enable container cpu usage collector.")
	rootCmd.Flags().BoolVar(&collectorContainerFS, "collector.container.fs", true, "Enable container fs collector.")
	rootCmd.Flags().BoolVar(&collectorContainerStats, "collector.container.stats", true, "Enable container stats collector.")
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

	if os.Getenv("IP") == "" {
		log.GetLogger().Info("IP environment variable not set, pass it do display the IP of the exporter on the homepage", "missing_env", "IP")
	}

	dockerClient, err := docker.NewDockerClient(dockerHost, sizeCacheDuration, diskUsageCacheDuration)
	if err != nil {
		log.GetLogger().Error("Failed to create Docker client", "error", err, "docker_host", dockerHost)
		os.Exit(1)
	}

	log.GetLogger().Info("Initializing Docker Prometheus exporter...")
	collectorConfig := exporter.CollectorConfig{
		System:           collectorSystem,
		Container:        collectorContainer || collectorContainerNetwork || collectorContainerFS || collectorContainerStats || collectorContainerCPU,
		ContainerNetwork: collectorContainerNetwork,
		ContainerCPU:     collectorContainerCPU,
		ContainerFS:      collectorContainerFS,
		ContainerStats:   collectorContainerStats,
	}
	var reg prometheus.Gatherer
	if internalMetrics {
		reg = prometheus.DefaultGatherer
		exporter.RegisterCollectorsWithRegistry(dockerClient, nil, Version, collectorConfig)
	} else {
		// Create a custom registry that doesn't include the Go collector, process collector, etc.
		registry := prometheus.NewRegistry()
		exporter.RegisterCollectorsWithRegistry(dockerClient, registry, Version, collectorConfig)
		reg = registry
	}

	registerHttp(dockerClient, reg)

	server := &http.Server{Addr: fmt.Sprintf("%s:%s", address, port), ErrorLog: slog.NewLogLogger(log.GetLogger().Handler(), slog.LevelWarn)}
	log.GetLogger().Info("HTTP server created")
	go func() {
		log.GetLogger().Info("Listening on metrics endpoint", "address", fmt.Sprintf("%s:%s", address, port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.GetLogger().Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Collect initial metrics in background
	go func() {
		log.GetLogger().Info("Collecting initial metrics in background...")
		ctx := context.Background()
		start := time.Now()

		// load disk usage cache
		dockerClient.Disk(ctx)

		// Perform an initial collection to warm up caches
		containers, err := dockerClient.ListAllRunningContainers(ctx)
		if err != nil {
			log.GetLogger().Warn("Initial container listing failed", "error", err)
		} else if len(containers) > 0 {
			// Warm up by inspecting and getting stats for all containers concurrently
			log.GetLogger().Debug("Warming up container stats cache", "count", len(containers))

			var wg sync.WaitGroup
			// Use a semaphore to limit concurrent requests
			sem := make(chan struct{}, 5)

			for _, container := range containers {
				wg.Add(1)
				go func(c docker.ContainerInfo) {
					defer wg.Done()
					sem <- struct{}{}        // acquire
					defer func() { <-sem }() // release

					_, _ = dockerClient.InspectContainer(ctx, c.ID, collectorConfig.ContainerFS)
					_, _ = dockerClient.GetContainerStats(ctx, c.ID, true) // always true because webapi shows cpu usage even if collector is disabled
				}(container)
			}
			wg.Wait()
		}

		// test slow startup
		//time.Sleep(10 * time.Second)

		duration := time.Since(start)
		log.GetLogger().Info(fmt.Sprintf("Initial metrics collection complete in %s", duration), "duration_ns", int(duration))
		glob.SetReady()
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

func registerHttp(dockerClient *docker.Client, reg prometheus.Gatherer) {
	http.Handle("/status", status.HandleStatus(dockerClient, Version))

	// Wrapper for /metrics that returns 503 when not ready
	metricsHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: trace,
	})
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !glob.IsReady() {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("503 Service Unavailable (Collecting initial metrics, please wait...)\n"))
			return
		}
		metricsHandler.ServeHTTP(w, r)
	}))

	// Web UI and API
	if homepage {
		http.HandleFunc("/", web.HandleRoot())
		http.HandleFunc("/main.css", web.HandleCss())
		http.HandleFunc("/main.js", web.HandleJs())
		http.HandleFunc("/chart.umd.min.js", web.HandleChartJs())

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
}
