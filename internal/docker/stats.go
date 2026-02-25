package docker

import (
	"context"
	"encoding/json"
	"io"

	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/client"
)

type ContainerCpuStats struct {
	// Raw CPU counters (ns).
	UsageNS          uint64
	UsageUserNS      uint64
	UsageKernelNS    uint64
	PreUsageNS       uint64
	SystemUsageNS    uint64
	PreSystemUsageNS uint64

	OnlineCpus uint32
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

type cpuEntry struct {
	UsageNS       uint64
	SystemUsageNS uint64
}

func (c *Client) GetContainerStats(ctx context.Context, containerID string, cpu bool) (ContainerStats, error) {
	stats, err := c.getContainerStats(ctx, containerID, cpu)
	if err != nil {
		glob.SetError("GetContainerStats", &err)
		return ContainerStats{}, err
	}
	glob.SetError("GetContainerStats", nil)
	return stats, nil
}

func (c *Client) getContainerStats(ctx context.Context, containerID string, cpu bool) (ContainerStats, error) {
	stats, err := c.client.ContainerStats(ctx, containerID, client.ContainerStatsOptions{
		Stream:                false,
		IncludePreviousSample: false,
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

	var data ContainerCpuStats
	if cpu {
		c.cpuStatsRWMutex.RLock()
		prev := c.cpuStatsCache[containerID]
		c.cpuStatsRWMutex.RUnlock()

		data = ContainerCpuStats{
			UsageNS:          rec.CpuStats.CpuUsage.TotalUsage,
			UsageUserNS:      rec.CpuStats.CpuUsage.UsageInUsermode,
			UsageKernelNS:    rec.CpuStats.CpuUsage.UsageInKernelmode,
			PreUsageNS:       prev.UsageNS,
			SystemUsageNS:    rec.CpuStats.SystemCpuUsage,
			PreSystemUsageNS: prev.SystemUsageNS,
			OnlineCpus:       rec.CpuStats.OnlineCpus,
		}

		c.cpuStatsRWMutex.Lock()
		c.cpuStatsCache[containerID] = cpuEntry{
			UsageNS:       rec.CpuStats.CpuUsage.TotalUsage,
			SystemUsageNS: rec.CpuStats.SystemCpuUsage,
		}
		c.cpuStatsRWMutex.Unlock()
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
		Cpu:              data,
		MemoryUsageKiB:   (rec.MemoryStats.Usage - rec.MemoryStats.Stats.InactiveFile) / 1024,
		MemoryLimitKiB:   rec.MemoryStats.Limit / 1024,
		Net:              net,
		BlockInputBytes:  blockInputBytes,
		BlockOutputBytes: blockOutputBytes,
	}

	return stat, nil
}
