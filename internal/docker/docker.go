package docker

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/h3rmt/docker-exporter/internal/log"

	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Client struct {
	client *client.Client

	// size cache for expensive ContainerList(Size:true)
	sizeMu          sync.Mutex
	sizeCache       map[string]sizeEntry // containerID -> sizes
	sizeLastUpdated time.Time
	sizeRefreshing  bool
	sizeRefreshCh   chan struct{}
}

func NewDockerClient(host string) (*Client, error) {
	c, err := client.New(
		client.WithHost(host),
		client.WithUserAgent("docker-exporter"),
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
	Ports       []container.PortSummary
	NetworkMode string
	Created     int64
	State       container.ContainerState
}

func (c *Client) ListAllRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := c.client.ContainerList(ctx, client.ContainerListOptions{
		All:  true,
		Size: false,
	})
	if err != nil {
		return nil, err
	}
	containerInfos := make([]ContainerInfo, len(containers.Items))
	for i, c := range containers.Items {

		names := make([]string, len(c.Names))
		for j, name := range c.Names {
			if len(name) > 0 && name[0] == '/' {
				names[j] = name[1:]
			} else {
				names[j] = name
			}
		}
		containerInfos[i] = ContainerInfo{
			ID:          c.ID,
			Names:       names,
			ImageID:     c.ImageID,
			Command:     c.Command,
			Ports:       c.Ports,
			NetworkMode: c.HostConfig.NetworkMode,
			State:       c.State,
			Created:     c.Created,
		}
		log.GetLogger().Log(ctx, log.LevelTrace, "Listed container", "container_id", containerInfos[i].ID, "names", containerInfos[i].Names, "state", containerInfos[i].State)
	}
	return containerInfos, nil
}

type ContainerInspect struct {
	ExitCode     int
	StartedAt    uint64
	FinishedAt   uint64
	RestartCount int
	SizeRootFs   int64
	SizeRw       int64
	NanoCpus     int64
}

type Inspect struct {
	RestartCount int              `json:"RestartCount"`
	State        *container.State `json:"State"`
	SizeRw       *int64           `json:"SizeRw,omitempty"`
	SizeRootFs   *int64           `json:"SizeRootFs,omitempty"`
	HostConfig   struct {
		// only this looks like the real value to be used for limiting cpu
		NanoCPUs  int64 `json:"NanoCpus"`  // CPU quota in units of 10<sup>-9</sup> CPUs.
		CPUShares int64 `json:"CpuShares"` // CPU shares (relative weight vs. other containers)

		// Applicable to UNIX platforms
		CPUPeriod          int64 `json:"CpuPeriod"`          // CPU CFS (Completely Fair Scheduler) period
		CPUQuota           int64 `json:"CpuQuota"`           // CPU CFS (Completely Fair Scheduler) quota
		CPURealtimePeriod  int64 `json:"CpuRealtimePeriod"`  // CPU real-time period
		CPURealtimeRuntime int64 `json:"CpuRealtimeRuntime"` // CPU real-time runtime

		// Applicable to Windows
		CPUCount   int64 `json:"CpuCount"`   // CPU count
		CPUPercent int64 `json:"CpuPercent"` // CPU percent
	} `json:"HostConfig"`
}

func (c *Client) InspectContainer(ctx context.Context, containerID string, size bool) (ContainerInspect, error) {
	var sizeRootFs int64
	var sizeRw int64
	if size {
		// Ensure we have current (or at least existing) cached size values.
		sizes := c.getCachedValues(ctx)
		// Apply cached sizes if available
		se, ok := sizes[containerID]
		if ok {
			sizeRootFs = se.SizeRootFs
			sizeRw = se.SizeRw
		}
	}

	inspect, err := c.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{
		Size: size && sizeRootFs == 0,
	})
	if err != nil {
		return ContainerInspect{}, err
	}
	var ret Inspect
	if err := json.Unmarshal(inspect.Raw, &ret); err != nil {
		return ContainerInspect{}, err
	}

	if size && sizeRootFs == 0 {
		sizeRootFs = *ret.SizeRootFs
		sizeRw = *ret.SizeRw
	}
	cInspect := ContainerInspect{
		ExitCode:     ret.State.ExitCode,
		StartedAt:    parseTimeOrEmpty(ret.State.StartedAt),
		FinishedAt:   parseTimeOrEmpty(ret.State.FinishedAt),
		RestartCount: ret.RestartCount,
		NanoCpus:     ret.HostConfig.NanoCPUs,
		SizeRootFs:   sizeRootFs,
		SizeRw:       sizeRw,
	}
	log.GetLogger().Log(ctx, log.LevelTrace, "Inspected container", "container_id", containerID, "exit_code", cInspect.ExitCode, "restart_count", cInspect.RestartCount)
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

type ContainerCpuStats struct {
	// Raw CPU counters (ns).
	UsageNS          uint64
	UsageUserNS      uint64
	UsageKernelNS    uint64
	PreUsageNS       uint64
	PreUsageUserNS   uint64
	PreUsageKernelNS uint64
	SystemUsageNS    uint64
	PreSystemUsageNS uint64
	OnlineCpus       uint32
}

type ContainerNetStats struct {
	SendBytes   uint64
	SendDropped uint64
	SendErrors  uint64
	RecvBytes   uint64
	RecvDropped uint64
	RecvErrors  uint64
}

type ContainerStats struct {
	PIds uint64

	Cpu ContainerCpuStats
	Net ContainerNetStats

	MemoryUsageKiB uint64
	MemoryLimitKiB uint64

	BlockInputBytes  uint64
	BlockOutputBytes uint64
}

type recStats struct {
	PidsStats struct {
		Current uint64 `json:"current"`
	} `json:"pids_stats"`
	CpuStats struct {
		SystemCpuUsage uint64 `json:"system_cpu_usage"`
		OnlineCpus     uint32 `json:"online_cpus"`
		CpuUsage       struct {
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
			TotalUsage        uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
	} `json:"cpu_stats"`
	PreCpuStats struct {
		SystemCpuUsage uint64 `json:"system_cpu_usage"`
		OnlineCpus     uint32 `json:"online_cpus"`
		CpuUsage       struct {
			UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64 `json:"usage_in_usermode"`
			TotalUsage        uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
	} `json:"precpu_stats"`
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

func (c *Client) GetContainerStats(ctx context.Context, containerID string) (ContainerStats, error) {
	stats, err := c.client.ContainerStats(ctx, containerID, client.ContainerStatsOptions{
		Stream:                false,
		IncludePreviousSample: true,
	})
	if err != nil {
		return ContainerStats{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "Failed to close container stats reader", "error", err)
		}
	}(stats.Body)

	var rec recStats
	if err := json.NewDecoder(stats.Body).Decode(&rec); err != nil {
		return ContainerStats{}, err
	}
	cpu := ContainerCpuStats{
		UsageNS:          rec.CpuStats.CpuUsage.TotalUsage,
		UsageUserNS:      rec.CpuStats.CpuUsage.UsageInUsermode,
		UsageKernelNS:    rec.CpuStats.CpuUsage.UsageInKernelmode,
		PreUsageNS:       rec.PreCpuStats.CpuUsage.TotalUsage,
		PreUsageUserNS:   rec.PreCpuStats.CpuUsage.UsageInUsermode,
		PreUsageKernelNS: rec.PreCpuStats.CpuUsage.UsageInKernelmode,
		SystemUsageNS:    rec.CpuStats.SystemCpuUsage,
		PreSystemUsageNS: rec.PreCpuStats.SystemCpuUsage,
		OnlineCpus:       rec.CpuStats.OnlineCpus,
	}

	// Network totals
	var netSendBytes uint64
	var netSendErrors uint64
	var netSendDropped uint64
	var netRecBytes uint64
	var netRecErrors uint64
	var netRecDropped uint64
	for _, net := range rec.Networks {
		netSendBytes += net.TxBytes
		netSendErrors += net.TxErrors
		netSendDropped += net.TxDropped
		netRecBytes += net.RxBytes
		netRecErrors += net.RxErrors
		netRecDropped += net.RxDropped
	}
	net := ContainerNetStats{
		SendBytes:   netSendBytes,
		SendDropped: netSendDropped,
		SendErrors:  netSendErrors,
		RecvBytes:   netRecBytes,
		RecvDropped: netRecDropped,
		RecvErrors:  netRecErrors,
	}

	// Block IO totals
	var blockInputBytes uint64
	var blockOutputBytes uint64
	for _, ioB := range rec.BlkioStats.IoServiceBytesRecursive {
		switch ioB.Op {
		case "read":
			blockInputBytes += uint64(ioB.Value)
		case "write":
			blockOutputBytes += uint64(ioB.Value)
		default:
			log.GetLogger().WarnContext(
				ctx,
				"Unknown blkio operation",
				"operation",
				ioB.Op,
				"container_id",
				containerID,
			)
		}
	}

	stat := ContainerStats{
		PIds:             rec.PidsStats.Current,
		Cpu:              cpu,
		MemoryUsageKiB:   (rec.MemoryStats.Usage - rec.MemoryStats.Stats.InactiveFile) / 1024,
		MemoryLimitKiB:   rec.MemoryStats.Limit / 1024,
		Net:              net,
		BlockInputBytes:  blockInputBytes,
		BlockOutputBytes: blockOutputBytes,
	}

	return stat, nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	data, err := c.client.Ping(ctx, client.PingOptions{})
	if err != nil {
		return "", err
	}
	return data.APIVersion, nil
}
