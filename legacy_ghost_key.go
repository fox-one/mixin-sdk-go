package mixin

import (
	"context"

	"github.com/fox-one/mixin-sdk-go/mixinnet"
)

type (
	// GhostKeys transaction ghost keys
	GhostKeys struct {
		Mask mixinnet.Key   `json:"mask"`
		Keys []mixinnet.Key `json:"keys"`
	}

	GhostInput struct {
		Receivers []string `json:"receivers"`
		Index     int      `json:"index"`
		Hint      string   `json:"hint"`
	}
)

func (c *Client) ReadGhostKeys(ctx context.Context, receivers []string, index int) (*GhostKeys, error) {
	input := &GhostInput{
		Receivers: receivers,
		Index:     index,
		Hint:      "",
	}

	var resp GhostKeys
	if err := c.Post(ctx, "/outputs", input, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) BatchReadGhostKeys(ctx context.Context, inputs []*GhostInput) ([]*GhostKeys, error) {
	var resp []*GhostKeys
	if err := c.Post(ctx, "/outputs", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
