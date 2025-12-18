package docker

import (
	"context"
	"docker-exporter/internal/log"
	"encoding/json"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"time"
)

type Client struct {
	client *client.Client
}

func NewDockerClient(host string) (*Client, error) {
	c, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

type ContainerInfo struct {
	ID          string
	Names       []string
	ImageID     string
	Command     string
	Ports       []container.Port
	NetworkMode string
	Created     int64
	State       container.ContainerState
	SizeRootFs  int64
	SizeRw      int64
}

func (client *Client) ListAllRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := client.client.ContainerList(ctx, container.ListOptions{All: true, Size: true})
	if err != nil {
		return nil, err
	}
	containerInfos := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		containerInfos[i] = ContainerInfo{
			ID:          c.ID,
			Names:       c.Names,
			ImageID:     c.ImageID,
			Command:     c.Command,
			Ports:       c.Ports,
			NetworkMode: c.HostConfig.NetworkMode,
			State:       c.State,
			Created:     c.Created,
			SizeRootFs:  c.SizeRootFs,
			SizeRw:      c.SizeRw,
		}
		log.DebugWith("Listed container", "container_id", containerInfos[i].ID, "names", containerInfos[i].Names, "state", containerInfos[i].State)
	}
	return containerInfos, nil
}

type ContainerInspect struct {
	ExitCode     int
	StartedAt    uint64
	FinishedAt   uint64
	RestartCount int
}

func (client *Client) InspectContainer(ctx context.Context, containerID string) (ContainerInspect, error) {
	inspect, err := client.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return ContainerInspect{}, err
	}
	cInspect := ContainerInspect{
		ExitCode:     inspect.State.ExitCode,
		StartedAt:    parseTimeOrEmpty(inspect.State.StartedAt),
		FinishedAt:   parseTimeOrEmpty(inspect.State.FinishedAt),
		RestartCount: inspect.RestartCount,
	}
	log.DebugWith("Inspected container", "container_id", containerID, "exit_code", cInspect.ExitCode, "restart_count", cInspect.RestartCount)
	return cInspect, nil
}

func parseTimeOrEmpty(data string) uint64 {
	zeroTime := "0001-01-01T00:00:00Z"
	if data == zeroTime {
		return 0
	}
	t, err := time.Parse(time.RFC3339, data)
	if err != nil {
		return 0
	}
	return uint64(t.Unix())
}

type ContainerStats struct {
	PIds                    uint64
	CPUinUserModeMicroSec   uint64 // microseconds  / 1000
	CPUinKernelModeMicroSec uint64 // microseconds  / 1000

	MemoryUsageKiB uint64 // / 1024 for KiB
	MemoryLimitKiB uint64 // / 1024 for KiB

	BlockOutputBytes uint64
	BlockInputBytes  uint64

	NetSendBytes   uint64
	NetSendErrors  uint64
	NetSendDropped uint64
	NetRecvBytes   uint64
	NetRecvErrors  uint64
	NetRecvDropped uint64
}

type recStats struct {
	PidsStats struct {
		Current uint64 `json:"current"`
	} `json:"pids_stats"`
	CpuStats struct {
		CpuUsage struct {
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
	} `json:"cpu_stats"`
	BlkioStats struct {
		IoServiceBytesRecursive []struct {
			Major int    `json:"major"`
			Minor int    `json:"minor"`
			Op    string `json:"op"`
			Value int    `json:"value"`
		} `json:"io_service_bytes_recursive"`
	} `json:"blkio_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
		Stats struct {
			//ActiveFile   uint64 `json:"active_file"`
			InactiveFile uint64 `json:"inactive_file"`
		} `json:"stats"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes   uint64 `json:"rx_bytes"`
		RxErrors  uint64 `json:"rx_errors"`
		RxDropped uint64 `json:"rx_dropped"`
		TxBytes   uint64 `json:"tx_bytes"`
		TxErrors  uint64 `json:"tx_errors"`
		TxDropped uint64 `json:"tx_dropped"`
	} `json:"networks"`
}

func (client *Client) GetContainerStats(ctx context.Context, containerID string) (ContainerStats, error) {
	stats, err := client.client.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return ContainerStats{}, err
	}
	defer stats.Body.Close()

	var recStats recStats
	err = json.NewDecoder(stats.Body).Decode(&recStats)
	if err != nil {
		return ContainerStats{}, err
	}

	var netSendBytes uint64
	var netSendErrors uint64
	var netSendDropped uint64
	var netRecBytes uint64
	var netRecErrors uint64
	var netRecDropped uint64
	for _, net := range recStats.Networks {
		netSendBytes += net.TxBytes
		netSendErrors += net.TxErrors
		netSendDropped += net.TxDropped
		netRecBytes += net.RxBytes
		netRecErrors += net.RxErrors
		netRecDropped += net.RxDropped
	}
	var blockInputBytes uint64
	var blockOutputBytes uint64
	for _, ioB := range recStats.BlkioStats.IoServiceBytesRecursive {
		if ioB.Op == "read" {
			blockInputBytes += uint64(ioB.Value)
		} else if ioB.Op == "write" {
			blockOutputBytes += uint64(ioB.Value)
		} else {
			log.WarningWith("Unknown blkio operation", "operation", ioB.Op, "container_id", containerID)
		}
	}

	log.DebugWith("Retrieved container stats", "container_id", containerID, "pids", recStats.PidsStats.Current, "mem_usage_kib", (recStats.MemoryStats.Usage-recStats.MemoryStats.Stats.InactiveFile)/1024)
	stat := ContainerStats{
		PIds:                    recStats.PidsStats.Current,
		CPUinUserModeMicroSec:   recStats.CpuStats.CpuUsage.UsageInUsermode / 1000,
		CPUinKernelModeMicroSec: recStats.CpuStats.CpuUsage.UsageInKernelmode / 1000,
		MemoryUsageKiB:          (recStats.MemoryStats.Usage - recStats.MemoryStats.Stats.InactiveFile) / 1024,
		MemoryLimitKiB:          recStats.MemoryStats.Limit / 1024,
		NetSendBytes:            netSendBytes,
		NetSendDropped:          netSendDropped,
		NetSendErrors:           netSendErrors,
		NetRecvBytes:            netRecBytes,
		NetRecvDropped:          netRecDropped,
		NetRecvErrors:           netRecErrors,
		BlockInputBytes:         blockInputBytes,
		BlockOutputBytes:        blockOutputBytes,
	}
	return stat, nil
}

func (client *Client) Ping(ctx context.Context) error {
	_, err := client.client.Ping(ctx)
	return err
}
