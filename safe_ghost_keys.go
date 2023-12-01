package mixin

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

type (
	// SafeGhostKeys transaction ghost keys
	SafeGhostKeys struct {
		Mask mixinnet.Key   `json:"mask"`
		Keys []mixinnet.Key `json:"keys"`
	}

	SafeGhostKeyInput struct {
		Receivers []string `json:"receivers"`
		Index     uint8    `json:"index"`
		Hint      string   `json:"hint"`
	}
)

func (c *Client) SafeCreateGhostKeys(ctx context.Context, inputs []*SafeGhostKeyInput) ([]*SafeGhostKeys, error) {
	var resp []*SafeGhostKeys
	if err := c.Post(ctx, "/safe/keys", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SafeCreateMixAddressGhostKeys(ctx context.Context, txVer uint8, trace string, ma *MixAddress, outputIndex uint8) (*SafeGhostKeys, error) {
	if len(ma.xinMembers) > 0 {
		r := mixinnet.GenerateKey(rand.Reader)
		gkr := &SafeGhostKeys{
			Mask: r.Public(),
			Keys: make([]mixinnet.Key, len(ma.xinMembers)),
		}
		for i, a := range ma.xinMembers {
			k := mixinnet.DeriveGhostPublicKey(txVer, &r, &a.PublicViewKey, &a.PublicSpendKey, outputIndex)
			gkr.Keys[i] = *k
		}
		return gkr, nil
	}

	gks, err := c.SafeCreateGhostKeys(ctx, []*SafeGhostKeyInput{
		{
			Receivers: ma.Members(),
			Index:     outputIndex,
			Hint:      uuidHash([]byte(fmt.Sprintf("trace:%s;index:%d", trace, outputIndex))),
		},
	})
	if err != nil {
		return nil, err
	}
	return gks[0], nil
}
