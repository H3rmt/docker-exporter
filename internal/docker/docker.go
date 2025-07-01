package docker

import (
	"context"
	"docker-exporter/internal/log"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func NewDockerClient(host string) (*client.Client, error) {
	return client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
}

type ContainerInfo struct {
	ID      string
	Names   []string
	ImageID string
	Command string
	Ports   []container.Port
	Created int64
	State   container.ContainerState
}

func ListAllRunningContainers(ctx context.Context, cli *client.Client) ([]ContainerInfo, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	containerInfos := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		containerInfos[i] = ContainerInfo{
			ID:      c.ID,
			Names:   c.Names,
			ImageID: c.ImageID,
			Command: c.Command,
			Ports:   c.Ports,
			State:   c.State,
			Created: c.Created,
		}
		log.Debug("Container: %v", containerInfos[i])
	}
	return containerInfos, nil
}
