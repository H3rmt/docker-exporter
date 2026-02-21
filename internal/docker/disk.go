package docker

import (
	"context"

	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/client"
)

type DiskUsage struct {
	ContainersTotalSize   int64
	ContainersReclaimable int64
	ImagesTotalSize       int64
	ImagesReclaimable     int64
	BuildCacheTotalSize   int64
	BuildCacheReclaimable int64
	VolumesTotalSize      int64
	VolumesReclaimable    int64
}

func loadDiskUsageFunction(c *client.Client) func(ctx context.Context) (DiskUsage, error) {
	return func(ctx context.Context) (DiskUsage, error) {
		data, err := c.DiskUsage(ctx, client.DiskUsageOptions{
			Containers: true,
			Images:     true,
			BuildCache: true,
			Volumes:    true,
			Verbose:    false,
		})
		if err != nil {
			glob.SetError("DiskUsage", &err)
			log.GetLogger().ErrorContext(ctx, "Failed to disk usage", "error", err)
			return DiskUsage{}, err
		}

		glob.SetError("DiskUsage", nil)
		return DiskUsage{
			ContainersTotalSize:   data.Containers.TotalSize,
			ContainersReclaimable: data.Containers.Reclaimable,
			ImagesTotalSize:       data.Images.TotalSize,
			ImagesReclaimable:     data.Images.Reclaimable,
			BuildCacheTotalSize:   data.BuildCache.TotalSize,
			BuildCacheReclaimable: data.BuildCache.Reclaimable,
			VolumesTotalSize:      data.Volumes.TotalSize,
			VolumesReclaimable:    data.Volumes.Reclaimable,
		}, nil
	}
}

func (c *Client) Disk(ctx context.Context) DiskUsage {
	data := c.diskUsageCache.GetValues(ctx)
	log.GetLogger().Log(ctx, log.LevelTrace, "disk usage cache", "data", data)
	return data
}

func (c *Client) disk(ctx context.Context) (DiskUsage, error) {
	data, err := c.client.DiskUsage(ctx, client.DiskUsageOptions{
		Containers: true,
		Images:     true,
		BuildCache: true,
		Volumes:    true,
		Verbose:    false,
	})
	if err != nil {
		return DiskUsage{}, err
	}
	return DiskUsage{
		ContainersTotalSize:   data.Containers.TotalSize,
		ContainersReclaimable: data.Containers.Reclaimable,
		ImagesTotalSize:       data.Images.TotalSize,
		ImagesReclaimable:     data.Images.Reclaimable,
		BuildCacheTotalSize:   data.BuildCache.TotalSize,
		BuildCacheReclaimable: data.BuildCache.Reclaimable,
		VolumesTotalSize:      data.Volumes.TotalSize,
		VolumesReclaimable:    data.Volumes.Reclaimable,
	}, nil
}
