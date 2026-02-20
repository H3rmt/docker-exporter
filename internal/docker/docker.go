package docker

import (
	"sync"
	"time"

	"github.com/h3rmt/docker-exporter/internal/glob"
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
		glob.SetError("NewDockerClient", &err)
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}
