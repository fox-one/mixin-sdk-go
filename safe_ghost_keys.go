package mixin

import (
	"context"

	"github.com/shopspring/decimal"
)

type (
	// GhostKeys transaction ghost keys
	GhostKeys struct {
		Mask Key   `json:"mask"`
		Keys []Key `json:"keys"`
	}

	GhostKeyInput struct {
		Receivers []string `json:"receivers"`
		Index     int      `json:"index"`
		Hint      string   `json:"hint"`
	}
)

func (g GhostKeys) DumpOutput(threshold uint8, amount decimal.Decimal) *Output {
	return &Output{
		Mask:   g.Mask,
		Keys:   g.Keys,
		Amount: NewIntegerFromDecimal(amount),
		Script: NewThresholdScript(threshold),
	}
}

func (c *Client) SafeCreateGhostKeys(ctx context.Context, inputs []*GhostKeyInput) ([]*GhostKeys, error) {
	var resp []*GhostKeys
	if err := c.Post(ctx, "/safe/keys", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
