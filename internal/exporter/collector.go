package exporter

import (
	"context"
	"docker-exporter/internal/docker"
	"docker-exporter/internal/log"
	"github.com/prometheus/client_golang/prometheus"
)

// DockerCollector implements the prometheus.Collector interface
type DockerCollector struct {
	dockerClient *docker.Client
	version      string
}

var (
	exporterInfoDesc = prometheus.NewDesc(
		"docker_exporter_info",
		"Information about the docker exporter",
		[]string{"version"},
		nil,
	)
	containerInfoDesc = prometheus.NewDesc(
		"docker_container_info",
		"Container information",
		[]string{"container_id", "name", "image_id", "command", "network_mode"},
		nil,
	)
	containerNameDesc = prometheus.NewDesc(
		"docker_container_name",
		"Name for the container (can be more than one)",
		[]string{"container_id", "name"},
		nil,
	)
	containerStateDesc = prometheus.NewDesc(
		"docker_container_state",
		"Container State (0=created, 1=running, 2=paused, 3=restarting, 4=removing, 5=exited, 6=dead)",
		[]string{"container_id"},
		nil,
	)
	containerCreatedDesc = prometheus.NewDesc(
		"docker_container_created_seconds",
		"Timestamp in seconds when the container was created",
		[]string{"container_id"},
		nil,
	)
	containerStartedDesc = prometheus.NewDesc(
		"docker_container_started_seconds",
		"Timestamp in seconds when the container was started",
		[]string{"container_id"},
		nil,
	)
	containerFinishedAtDesc = prometheus.NewDesc(
		"docker_container_finished_at_seconds",
		"Timestamp in seconds when the container finished",
		[]string{"container_id"},
		nil,
	)
	containerPortsDesc = prometheus.NewDesc(
		"docker_container_ports",
		"Forwarded Ports",
		[]string{"container_id", "public_port", "private_port", "ip", "type"},
		nil,
	)
	containerSizeRootFsDesc = prometheus.NewDesc(
		"docker_container_rootfs_size_bytes",
		"Size of rootfs in this container in bytes",
		[]string{"container_id"},
		nil,
	)
	containerSizeRwDesc = prometheus.NewDesc(
		"docker_container_rw_size_bytes",
		"Size of files that have been created or changed by this container in bytes",
		[]string{"container_id"},
		nil,
	)
	containerPidsDesc = prometheus.NewDesc(
		"docker_container_pids",
		"Number of processes running in the container",
		[]string{"container_id"},
		nil,
	)
	containerCpuUserMicroSecondsDesc = prometheus.NewDesc(
		"docker_container_cpu_user_microseconds_total",
		"Time (in microseconds) spent by tasks in user mode",
		[]string{"container_id"},
		nil,
	)
	containerCpuKernelMicroSecondsDesc = prometheus.NewDesc(
		"docker_container_cpu_kernel_microseconds_total",
		"Time (in microseconds) spent by tasks in kernel mode",
		[]string{"container_id"},
		nil,
	)
	containerMemLimitKiBDesc = prometheus.NewDesc(
		"docker_container_mem_limit_kib",
		"Container memory limit in KiB",
		[]string{"container_id"},
		nil,
	)
	containerMemUsageKiBDesc = prometheus.NewDesc(
		"docker_container_mem_usage_kib",
		"Container memory usage in KiB",
		[]string{"container_id"},
		nil,
	)
	containerNetSendBytesDesc = prometheus.NewDesc(
		"docker_container_net_send_bytes_total",
		"Total number of bytes sent",
		[]string{"container_id"},
		nil,
	)
	containerNetSendDroppedDesc = prometheus.NewDesc(
		"docker_container_net_send_dropped_total",
		"Total number of send packet drop",
		[]string{"container_id"},
		nil,
	)
	containerNetSendErrorsDesc = prometheus.NewDesc(
		"docker_container_net_send_errors_total",
		"Total number of send errors",
		[]string{"container_id"},
		nil,
	)
	containerNetRecvBytesDesc = prometheus.NewDesc(
		"docker_container_net_receive_bytes_total",
		"Total number of bytes received",
		[]string{"container_id"},
		nil,
	)
	containerNetRecvDroppedDesc = prometheus.NewDesc(
		"docker_container_net_receive_dropped_total",
		"Total number of receive packet drop",
		[]string{"container_id"},
		nil,
	)
	containerNetRecvErrorsDesc = prometheus.NewDesc(
		"docker_container_net_receive_errors_total",
		"Total number of receive errors",
		[]string{"container_id"},
		nil,
	)
	containerBlockInputBytesDesc = prometheus.NewDesc(
		"docker_container_block_input_total",
		"Total number of bytes read from disk",
		[]string{"container_id"},
		nil,
	)
	containerBlockOutputBytesDesc = prometheus.NewDesc(
		"docker_container_block_output_total",
		"Total number of bytes written to disk",
		[]string{"container_id"},
		nil,
	)
	containerExitCodeDesc = prometheus.NewDesc(
		"docker_container_exit_code",
		"Exit code of the container",
		[]string{"container_id"},
		nil,
	)
	containerRestartCountDesc = prometheus.NewDesc(
		"docker_container_restart_count",
		"Number of times the container has been restarted",
		[]string{"container_id"},
		nil,
	)
)

func NewDockerCollector(client *docker.Client, version string) *DockerCollector {
	return &DockerCollector{
		dockerClient: client,
		version:      version,
	}
}

func (c *DockerCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *DockerCollector) Collect(ch chan<- prometheus.Metric) {
	// Export version information
	ch <- prometheus.MustNewConstMetric(
		exporterInfoDesc,
		prometheus.GaugeValue,
		1,
		c.version,
	)

	ctx := context.Background()

	containerInfo, err := c.dockerClient.ListAllRunningContainers(ctx)
	if err != nil {
		log.Warning("Failed to list running containers: %v", err)
	} else {
		formatContainerInfo(ch, containerInfo)
		formatContainerNames(ch, containerInfo)
		formatContainerState(ch, containerInfo)
		formatContainerCreated(ch, containerInfo)
		formatContainerPorts(ch, containerInfo)
		formatContainerSizeRootFs(ch, containerInfo)
		formatContainerSizeRw(ch, containerInfo)
	}

	for _, container := range containerInfo {
		id := container.ID
		inspect, err := c.dockerClient.InspectContainer(ctx, id)
		if err != nil {
			log.Warning("Failed to inspect container %s: %v", id, err)
		} else {
			formatContainerStarted(ch, id, inspect)
			formatContainerExitCode(ch, id, inspect)
			formatContainerRestartCount(ch, id, inspect)
			formatContainerFinished(ch, id, inspect)
		}
	}

	for _, container := range containerInfo {
		id := container.ID
		stat, err := c.dockerClient.GetContainerStats(ctx, id)
		if err != nil {
			log.Warning("Failed to get container stats for container %s: %v", id, err)
		} else {
			formatContainerPids(ch, id, stat)
			formatContainerCpuUserMicroSeconds(ch, id, stat)
			formatContainerCpuKernelMicroSeconds(ch, id, stat)
			formatContainerMemLimitKiB(ch, id, stat)
			formatContainerMemUsageKiB(ch, id, stat)
			formatContainerNetSendBytes(ch, id, stat)
			formatContainerNetSendDropped(ch, id, stat)
			formatContainerNetSendErrors(ch, id, stat)
			formatContainerNetRecvBytes(ch, id, stat)
			formatContainerNetRecvDropped(ch, id, stat)
			formatContainerNetRecvErrors(ch, id, stat)
			formatBlockOutputBytes(ch, id, stat)
			formatBlockInputBytes(ch, id, stat)
		}
	}
}

func RegisterCollectorsWithRegistry(cli *docker.Client, reg *prometheus.Registry, version string) {
	collector := NewDockerCollector(cli, version)
	if reg == nil {
		prometheus.MustRegister(collector)
	} else {
		reg.MustRegister(collector)
	}
}
