package exporter

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func NewDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

func ListRunningContainers(cli *client.Client) (int, error) {
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return 0, err
	}
	return len(containers), nil
}
