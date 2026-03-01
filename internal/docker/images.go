package docker

import (
	"context"

	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/client"
)

type ImageInfo struct {
	ID         string
	Containers int64
	Name       string
	Size       int64
	Created    int64
}

func (c *Client) ListAllImages(ctx context.Context) ([]ImageInfo, error) {
	images, err := c.listAllImages(ctx)
	if err != nil {
		glob.SetError("ListAllImages", &err)
		return nil, err
	}
	glob.SetError("ListAllImages", nil)
	return images, nil
}

func (c *Client) listAllImages(ctx context.Context) ([]ImageInfo, error) {
	images, err := c.client.ImageList(ctx, client.ImageListOptions{
		All:        false,
		SharedSize: false,
	})
	if err != nil {
		return nil, err
	}
	imageInfos := make([]ImageInfo, len(images.Items))
	for i, c := range images.Items {
		name := ""
		if len(c.RepoTags) > 0 {
			name = c.RepoTags[0]
		}
		imageInfos[i] = ImageInfo{
			ID:         c.ID,
			Containers: c.Containers,
			Name:       name,
			Size:       c.Size,
			Created:    c.Created,
		}
		log.GetLogger().Log(ctx, log.LevelTrace, "Found image", "id", c.ID, "containers", c.Containers, "size", c.Size, "created", c.Created, "name", name)
	}

	log.GetLogger().DebugContext(ctx, "Found images", "count", len(imageInfos))
	return imageInfos, nil
}
