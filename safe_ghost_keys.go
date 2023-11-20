package mixin

import (
	"context"
	"crypto/rand"

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

func (c *Client) SafeCreateMixAddressGhostKeys(ctx context.Context, trace string, ma *MixAddress, outputIndex uint) (*GhostKeys, error) {
	if len(ma.xinMembers) > 0 {
		r := NewKey(rand.Reader)
		gkr := &GhostKeys{
			Mask: r.Public(),
			Keys: make([]Key, len(ma.xinMembers)),
		}
		for i, a := range ma.xinMembers {
			k := DeriveGhostPublicKey(&r, &a.PublicViewKey, &a.PublicSpendKey, uint64(outputIndex))
			gkr.Keys[i] = *k
		}
		return gkr, nil
	}

	gks, err := c.SafeCreateGhostKeys(ctx, []*GhostKeyInput{
		{
			Receivers: ma.Members(),
			Index:     int(ma.Threshold),
			Hint:      trace,
		},
	})
	if err != nil {
		return nil, err
	}
	return gks[0], nil
}
