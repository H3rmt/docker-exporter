package docker

import (
	"time"

	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/moby/moby/client"
)

type Client struct {
	client *client.Client

	// size cache for expensive ContainerList(Size:true)
	sizeCache Cache[map[string]sizeEntry] // containerID -> sizes

	// disk usage cache for expensive DiskUsage
	diskUsageCache Cache[DiskUsage]
}

func NewDockerClient(host string, sizeCacheDuration time.Duration, diskUsageCacheDuration time.Duration) (*Client, error) {
	c, err := client.New(
		client.WithHost(host),
		client.WithUserAgent("docker-exporter"),
	)
	if err != nil {
		glob.SetError("NewDockerClient", &err)
		return nil, err
	}
	return &Client{
		client:         c,
		sizeCache:      NewCacheFull("sizeCache", sizeCacheDuration, loadContainerSizeFunction(c), copyMap),
		diskUsageCache: NewCache("diskUsageCache", diskUsageCacheDuration, loadDiskUsageFunction(c)),
	}, nil
}
