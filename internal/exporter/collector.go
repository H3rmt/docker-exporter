package exporter

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/h3rmt/docker-exporter/internal/osinfo"

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
		[]string{"hostname", "version"},
		nil,
	)
	hostOSInfoDesc = prometheus.NewDesc(
		"docker_exporter_host_os_info",
		"Information about the host operating system",
		[]string{"hostname", "os_name", "os_version"},
		nil,
	)
	containerInfoDesc = prometheus.NewDesc(
		"docker_container_info",
		"Container information",
		[]string{"hostname", "container_id", "name", "image_id", "command", "network_mode"},
		nil,
	)
	containerNameDesc = prometheus.NewDesc(
		"docker_container_name",
		"Name for the container (can be more than one)",
		[]string{"hostname", "container_id", "name"},
		nil,
	)
	containerStateDesc = prometheus.NewDesc(
		"docker_container_state",
		"Container State (0=created, 1=running, 2=paused, 3=restarting, 4=removing, 5=exited, 6=dead)",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerCreatedDesc = prometheus.NewDesc(
		"docker_container_created_seconds",
		"Timestamp in seconds when the container was created",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerPortsDesc = prometheus.NewDesc(
		"docker_container_ports",
		"Forwarded Ports",
		[]string{"hostname", "container_id", "public_port", "private_port", "ip", "type"},
		nil,
	)
	containerStartedDesc = prometheus.NewDesc(
		"docker_container_started_seconds",
		"Timestamp in seconds when the container was started",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerFinishedAtDesc = prometheus.NewDesc(
		"docker_container_finished_at_seconds",
		"Timestamp in seconds when the container finished",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerSizeRootFsDesc = prometheus.NewDesc(
		"docker_container_rootfs_size_bytes",
		"Size of rootfs in this container in bytes",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerSizeRwDesc = prometheus.NewDesc(
		"docker_container_rw_size_bytes",
		"Size of files that have been created or changed by this container in bytes",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerRestartCountDesc = prometheus.NewDesc(
		"docker_container_restart_count",
		"Number of times the container has been restarted",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerExitCodeDesc = prometheus.NewDesc(
		"docker_container_exit_code",
		"Exit code of the container",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerPidsDesc = prometheus.NewDesc(
		"docker_container_pids",
		"Number of processes running in the container",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerCpuUserNSDesc = prometheus.NewDesc(
		"docker_container_cpu_user_nanoseconds_total",
		"Time (in nanoseconds) spent by tasks in user mode",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerCpuKernelNSDesc = prometheus.NewDesc(
		"docker_container_cpu_kernel_nanoseconds_total",
		"Time (in nanoseconds) spent by tasks in kernel mode",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerCpuNSDesc = prometheus.NewDesc(
		"docker_container_cpu_nanoseconds_total",
		"Time (in nanoseconds) spent by tasks",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerCpuPercent = prometheus.NewDesc(
		"docker_container_cpu_percent",
		"Percentage of CPU used by the container (relative to max available CPU cores)",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerCpuPercentHost = prometheus.NewDesc(
		"docker_container_cpu_percent_host",
		"Percentage of CPU used by the container (relative to host CPU cores)",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerMemLimitKiBDesc = prometheus.NewDesc(
		"docker_container_mem_limit_kib",
		"Container memory limit in KiB",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerMemUsageKiBDesc = prometheus.NewDesc(
		"docker_container_mem_usage_kib",
		"Container memory usage in KiB",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerNetSendBytesDesc = prometheus.NewDesc(
		"docker_container_net_send_bytes_total",
		"Total number of bytes sent",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerNetSendDroppedDesc = prometheus.NewDesc(
		"docker_container_net_send_dropped_total",
		"Total number of send packet drop",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerNetSendErrorsDesc = prometheus.NewDesc(
		"docker_container_net_send_errors_total",
		"Total number of send errors",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerNetRecvBytesDesc = prometheus.NewDesc(
		"docker_container_net_receive_bytes_total",
		"Total number of bytes received",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerNetRecvDroppedDesc = prometheus.NewDesc(
		"docker_container_net_receive_dropped_total",
		"Total number of receive packet drop",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerNetRecvErrorsDesc = prometheus.NewDesc(
		"docker_container_net_receive_errors_total",
		"Total number of receive errors",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerBlockInputBytesDesc = prometheus.NewDesc(
		"docker_container_block_input_total",
		"Total number of bytes read from disk",
		[]string{"hostname", "container_id"},
		nil,
	)
	containerBlockOutputBytesDesc = prometheus.NewDesc(
		"docker_container_block_output_total",
		"Total number of bytes written to disk",
		[]string{"hostname", "container_id"},
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

type containerResult struct {
	id      string
	inspect docker.ContainerInspect
	stat    docker.ContainerStats
}

func (c *DockerCollector) Collect(ch chan<- prometheus.Metric) {
	hostname := getHostname()
	// Export version information
	ch <- prometheus.MustNewConstMetric(
		exporterInfoDesc,
		prometheus.GaugeValue,
		1,
		hostname,
		c.version,
	)
	// Export OS information
	osInfo := osinfo.GetCached()
	ch <- prometheus.MustNewConstMetric(
		hostOSInfoDesc,
		prometheus.GaugeValue,
		1,
		hostname,
		osInfo.Name,
		osInfo.VersionID,
	)
	ctx := context.Background()

	containerInfo, err := c.dockerClient.ListAllRunningContainers(ctx)
	if err != nil {
		log.GetLogger().WarnContext(ctx, "Failed to list running containers", "error", err)
		return
	}

	formatContainerInfo(ch, hostname, containerInfo)
	formatContainerNames(ch, hostname, containerInfo)
	formatContainerState(ch, hostname, containerInfo)
	formatContainerCreated(ch, hostname, containerInfo)
	formatContainerPorts(ch, hostname, containerInfo)

	resultCh := make(chan containerResult, len(containerInfo))
	var wg sync.WaitGroup

	for _, container := range containerInfo {
		wg.Add(1)
		go func(container docker.ContainerInfo) {
			defer wg.Done()
			id := container.ID

			inspect, err := c.dockerClient.InspectContainer(ctx, id, true)
			if err != nil {
				log.GetLogger().WarnContext(ctx, "Failed to inspect container", "error", err, "container_id", id)
			} else {
				stat, err := c.dockerClient.GetContainerStats(ctx, id)
				if err != nil {
					log.GetLogger().WarnContext(ctx, "Failed to get container stats", "error", err, "container_id", id)
				} else {
					result := containerResult{id: id, stat: stat, inspect: inspect}
					resultCh <- result
				}
			}

		}(container)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		formatContainerStarted(ch, hostname, result.id, result.inspect)
		formatContainerExitCode(ch, hostname, result.id, result.inspect)
		formatContainerRestartCount(ch, hostname, result.id, result.inspect)
		formatContainerFinished(ch, hostname, result.id, result.inspect)
		formatContainerSizeRootFs(ch, hostname, result.id, result.inspect)
		formatContainerSizeRw(ch, hostname, result.id, result.inspect)
		formatContainerPids(ch, hostname, result.id, result.stat)
		formatContainerCpuMicroSeconds(ch, hostname, result.id, result.stat)
		formatContainerCpuUserMicroSeconds(ch, hostname, result.id, result.stat)
		formatContainerCpuKernelMicroSeconds(ch, hostname, result.id, result.stat)
		formatContainerCpuPercentHost(ch, hostname, result.id, result.stat)
		formatContainerCpuPercent(ch, hostname, result.id, result.stat, result.inspect)
		formatContainerMemLimitKiB(ch, hostname, result.id, result.stat)
		formatContainerMemUsageKiB(ch, hostname, result.id, result.stat)
		formatContainerNetSendBytes(ch, hostname, result.id, result.stat)
		formatContainerNetSendDropped(ch, hostname, result.id, result.stat)
		formatContainerNetSendErrors(ch, hostname, result.id, result.stat)
		formatContainerNetRecvBytes(ch, hostname, result.id, result.stat)
		formatContainerNetRecvDropped(ch, hostname, result.id, result.stat)
		formatContainerNetRecvErrors(ch, hostname, result.id, result.stat)
		formatBlockOutputBytes(ch, hostname, result.id, result.stat)
		formatBlockInputBytes(ch, hostname, result.id, result.stat)
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

func getHostname() string {
	hn, _ := os.ReadFile("/etc/hostname")
	return strings.TrimSpace(string(hn))
}
