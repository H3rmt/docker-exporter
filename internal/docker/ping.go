package docker

import (
	"context"

	"github.com/moby/moby/client"
)

func (c *Client) Ping(ctx context.Context) (string, error) {
	data, err := c.client.Ping(ctx, client.PingOptions{})
	if err != nil {
		return "", err
	}
	return data.APIVersion, nil
}
