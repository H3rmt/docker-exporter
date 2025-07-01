package exporter

import (
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	runningContainers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "docker_running_containers",
		Help: "Number of running Docker containers",
	})
)

// RegisterCollectorsWithRegistry registers metrics with a custom registry
func RegisterCollectorsWithRegistry(cli *client.Client, reg *prometheus.Registry) {
	if reg == nil {
		prometheus.MustRegister(runningContainers)
	} else {
		reg.MustRegister(runningContainers)
	}
}
