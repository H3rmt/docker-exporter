package docker

import (
	"context"

	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/client"
)

func (c *Client) ListAllRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := c.listAllRunningContainers(ctx)
	if err != nil {
		glob.SetError("ListAllRunningContainers", &err)
		return nil, err
	}
	glob.SetError("ListAllRunningContainers", nil)
	return containers, nil
}

func (c *Client) listAllRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
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
