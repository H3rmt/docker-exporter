package docker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

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
	inspect, err := c.inspectContainer(ctx, containerID, size)
	if err != nil {
		glob.SetError("InspectContainer", &err)
		return ContainerInspect{}, err
	}
	glob.SetError("InspectContainer", nil)
	return inspect, nil
}

type sizeEntry struct {
	SizeRootFs int64
	SizeRw     int64
}

func loadContainerSizeFunction(c *client.Client) func(ctx context.Context) (map[string]sizeEntry, error) {
	return func(ctx context.Context) (map[string]sizeEntry, error) {
		containers, err := c.ContainerList(ctx, client.ContainerListOptions{
			All:  true,
			Size: true,
		})
		// Prepare results
		sizes := make(map[string]sizeEntry)
		if err != nil {
			glob.SetError("refreshSizes", &err)
			log.GetLogger().ErrorContext(ctx, "Failed to refresh container sizes", "error", err)
		} else {
			glob.SetError("refreshSizes", nil)
			for _, item := range containers.Items {
				sizes[item.ID] = sizeEntry{SizeRootFs: item.SizeRootFs, SizeRw: item.SizeRw}
			}
		}
		return sizes, nil
	}
}

func (c *Client) inspectContainer(ctx context.Context, containerID string, size bool) (ContainerInspect, error) {
	var sizeRootFs int64
	var sizeRw int64
	if size {
		// Ensure we have current (or at least existing) cached size values.
		sizes := c.sizeCache.GetValues(ctx)
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
