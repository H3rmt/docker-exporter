package exporter

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/h3rmt/docker-exporter/internal/osinfo"

	"github.com/prometheus/client_golang/prometheus"
)

// CollectorConfig holds which collector groups are enabled
type CollectorConfig struct {
	System           bool
	Container        bool
	ContainerNetwork bool
	ContainerFS      bool
	ContainerStats   bool
}

// DockerCollector implements the prometheus.Collector interface
type DockerCollector struct {
	dockerClient *docker.Client
	version      string
	config       CollectorConfig
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
	dockerDiskUsageContainersTotalSize = prometheus.NewDesc(
		"docker_disk_usage_container_total_size",
		"Information about Size of containers on disk.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageContainersReclaimable = prometheus.NewDesc(
		"docker_disk_usage_container_reclaimable",
		"Information about Size of containers on disk that can be reclaimed.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageImagesTotalSize = prometheus.NewDesc(
		"docker_disk_usage_images_total_size",
		"Information about Size of images on disk.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageImagesReclaimable = prometheus.NewDesc(
		"docker_disk_usage_images_reclaimable",
		"Information about Size of images on disk that can be reclaimed.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageBuildCacheTotalSize = prometheus.NewDesc(
		"docker_disk_usage_build_cache_total_size",
		"Information about Size of build cache on disk.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageBuildCacheReclaimable = prometheus.NewDesc(
		"docker_disk_usage_build_cache_reclaimable",
		"Information about Size of build on disk that can be reclaimed.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageVolumesTotalSize = prometheus.NewDesc(
		"docker_disk_usage_volumes_total_size",
		"Information about Size of volumes on disk.",
		[]string{"hostname"},
		nil,
	)
	dockerDiskUsageVolumesReclaimable = prometheus.NewDesc(
		"docker_disk_usage_volumes_reclaimable",
		"Information about Size of volumes on disk that can be reclaimed.",
		[]string{"hostname"},
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

func NewDockerCollector(client *docker.Client, version string, config CollectorConfig) *DockerCollector {
	return &DockerCollector{
		dockerClient: client,
		version:      version,
		config:       config,
	}
}

func (c *DockerCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.config.System {
		for _, desc := range []*prometheus.Desc{
			exporterInfoDesc, hostOSInfoDesc,
			dockerDiskUsageContainersTotalSize, dockerDiskUsageContainersReclaimable,
			dockerDiskUsageImagesTotalSize, dockerDiskUsageImagesReclaimable,
			dockerDiskUsageBuildCacheTotalSize, dockerDiskUsageBuildCacheReclaimable,
			dockerDiskUsageVolumesTotalSize, dockerDiskUsageVolumesReclaimable,
		} {
			ch <- desc
		}
	}

	if c.config.Container {
		for _, desc := range []*prometheus.Desc{
			containerInfoDesc, containerNameDesc, containerStateDesc, containerCreatedDesc,
			containerPortsDesc, containerStartedDesc, containerFinishedAtDesc,
			containerRestartCountDesc, containerExitCodeDesc,
		} {
			ch <- desc
		}
	}

	if c.config.ContainerStats {
		for _, desc := range []*prometheus.Desc{
			containerPidsDesc,
			containerCpuUserNSDesc, containerCpuKernelNSDesc, containerCpuNSDesc, containerCpuPercent, containerCpuPercentHost,
			containerMemLimitKiBDesc, containerMemUsageKiBDesc,
			containerBlockInputBytesDesc, containerBlockOutputBytesDesc,
		} {
			ch <- desc
		}
	}

	if c.config.ContainerNetwork {
		for _, desc := range []*prometheus.Desc{
			containerNetSendBytesDesc, containerNetSendDroppedDesc, containerNetSendErrorsDesc,
			containerNetRecvBytesDesc, containerNetRecvDroppedDesc, containerNetRecvErrorsDesc,
		} {
			ch <- desc
		}
	}

	if c.config.ContainerFS {
		ch <- containerSizeRootFsDesc
		ch <- containerSizeRwDesc
	}
}

type containerResult struct {
	id      string
	inspect docker.ContainerInspect
	stat    docker.ContainerStats
}

func (c *DockerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	start := time.Now()

	hostname := getHostname(ctx)

	if c.config.System {
		formatSystemInfo(ch, hostname, c.version)
		osInfo := osinfo.GetOSInfo(ctx)
		formatSystemHostInfo(ch, hostname, osInfo)
		disk := c.dockerClient.Disk(ctx)
		formatSystemDiskInfo(ch, hostname, disk)
	}
	log.GetLogger().DebugContext(ctx, "Finished collecting system metrics", "time", time.Since(start))

	if !c.config.Container {
		return
	}

	containerInfo, err := c.dockerClient.ListAllRunningContainers(ctx)
	if err != nil {
		log.GetLogger().ErrorContext(ctx, "Failed to list running containers", "error", err)
		return
	}
	log.GetLogger().DebugContext(ctx, "Found running containers", "time", time.Since(start), "count", len(containerInfo))

	formatContainerInfo(ch, hostname, containerInfo)
	formatContainerNames(ch, hostname, containerInfo)
	formatContainerState(ch, hostname, containerInfo)
	formatContainerCreated(ch, hostname, containerInfo)
	formatContainerPorts(ch, hostname, containerInfo)

	needStat := c.config.ContainerStats || c.config.ContainerNetwork

	resultCh := make(chan containerResult, len(containerInfo))
	var wg sync.WaitGroup

	for _, container := range containerInfo {
		wg.Add(1)
		go func(container docker.ContainerInfo) {
			defer wg.Done()
			id := container.ID

			inspect, err := c.dockerClient.InspectContainer(ctx, id, c.config.ContainerFS)
			if err != nil {
				log.GetLogger().WarnContext(ctx, "Failed to inspect container", "error", err, "container_id", id)
				return
			}

			var stat docker.ContainerStats
			if needStat {
				stat, err = c.dockerClient.GetContainerStats(ctx, id)
				if err != nil {
					log.GetLogger().WarnContext(ctx, "Failed to get container stats", "error", err, "container_id", id)
					return
				}
			}

			resultCh <- containerResult{id: id, stat: stat, inspect: inspect}
		}(container)
	}
	log.GetLogger().DebugContext(ctx, "Waiting for container stats", "time", time.Since(start), "count", len(containerInfo))

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		if c.config.Container {
			formatContainerStarted(ch, hostname, result.id, result.inspect)
			formatContainerFinished(ch, hostname, result.id, result.inspect)
			formatContainerExitCode(ch, hostname, result.id, result.inspect)
			formatContainerRestartCount(ch, hostname, result.id, result.inspect)
		}

		if c.config.ContainerFS {
			formatContainerSizeRootFs(ch, hostname, result.id, result.inspect)
			formatContainerSizeRw(ch, hostname, result.id, result.inspect)
		}

		if c.config.ContainerStats {
			formatContainerPids(ch, hostname, result.id, result.stat)
			formatContainerCpuMicroSeconds(ch, hostname, result.id, result.stat)
			formatContainerCpuUserMicroSeconds(ch, hostname, result.id, result.stat)
			formatContainerCpuKernelMicroSeconds(ch, hostname, result.id, result.stat)
			formatContainerCpuPercentHost(ch, hostname, result.id, result.stat)
			formatContainerCpuPercent(ch, hostname, result.id, result.stat, result.inspect)
			formatContainerMemLimitKiB(ch, hostname, result.id, result.stat)
			formatContainerMemUsageKiB(ch, hostname, result.id, result.stat)
			formatBlockOutputBytes(ch, hostname, result.id, result.stat)
			formatBlockInputBytes(ch, hostname, result.id, result.stat)
		}

		if c.config.ContainerNetwork {
			formatContainerNetSendBytes(ch, hostname, result.id, result.stat)
			formatContainerNetSendDropped(ch, hostname, result.id, result.stat)
			formatContainerNetSendErrors(ch, hostname, result.id, result.stat)
			formatContainerNetRecvBytes(ch, hostname, result.id, result.stat)
			formatContainerNetRecvDropped(ch, hostname, result.id, result.stat)
			formatContainerNetRecvErrors(ch, hostname, result.id, result.stat)
		}
	}

	log.GetLogger().DebugContext(ctx, "Finished collecting metrics", "time", time.Since(start))
}

func RegisterCollectorsWithRegistry(cli *docker.Client, reg *prometheus.Registry, version string, config CollectorConfig) {
	collector := NewDockerCollector(cli, version, config)
	if reg == nil {
		prometheus.MustRegister(collector)
	} else {
		reg.MustRegister(collector)
	}
}

func getHostname(ctx context.Context) string {
	hn, err := os.ReadFile("/etc/hostname")
	if err != nil {
		log.GetLogger().ErrorContext(ctx, "failed to read hostname", "error", err)
		glob.SetError("readHostname", &err)
		return ""
	}
	glob.SetError("readHostname", nil)
	return strings.TrimSpace(string(hn))
}
