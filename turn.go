package mixin

import (
	"context"
)

type Turn struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	Credential string `json:"credential"`
}

func (c *Client) ReadTurnServers(ctx context.Context) ([]*Turn, error) {
	var servers []*Turn
	if err := c.Get(ctx, "/turn", nil, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}
