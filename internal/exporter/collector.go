package exporter

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

// DockerCollector implements the prometheus.Collector interface
type DockerCollector struct {
	dockerClient *client.Client
}

var (
	containerInfoDesc = prometheus.NewDesc(
		"docker_container_info",
		"Docker container info",
		[]string{"container_id", "name", "image_id", "command"},
		nil,
	)
	containerNameDesc = prometheus.NewDesc(
		"docker_container_name",
		"Docker container name",
		[]string{"container_id", "name"},
		nil,
	)
	containerStateDesc = prometheus.NewDesc(
		"docker_container_state",
		"Docker container state (0=created, 1=running, 2=paused, 3=restarting, 4=removing, 5=exited, 6=dead)",
		[]string{"container_id"},
		nil,
	)
	containerCreatedDesc = prometheus.NewDesc(
		"docker_container_created",
		"Docker container created timestamp",
		[]string{"container_id"},
		nil,
	)
)

// NewDockerCollector creates a new DockerCollector
func NewDockerCollector(cli *client.Client) *DockerCollector {
	return &DockerCollector{
		dockerClient: cli,
	}
}

// Describe implements the prometheus.Collector interface
func (c *DockerCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements the prometheus.Collector interface
func (c *DockerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	containerInfo, err := ListAllRunningContainers(ctx, c.dockerClient)
	if err != nil {
		log.Printf("Failed to list running containers: %v", err)
	} else {
		formatContainerInfo(ch, containerInfo)
		formatContainerNames(ch, containerInfo)
		formatContainerState(ch, containerInfo)
		formatContainerCreated(ch, containerInfo)
	}
}

// RegisterCollectorsWithRegistry registers metrics with a custom registry
func RegisterCollectorsWithRegistry(cli *client.Client, reg *prometheus.Registry) {
	collector := NewDockerCollector(cli)
	if reg == nil {
		prometheus.MustRegister(collector)
	} else {
		reg.MustRegister(collector)
	}
}
