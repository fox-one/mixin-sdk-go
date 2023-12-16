package mixin

import (
	"context"
	"crypto/rand"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

func (c *Client) SafeCreateGhostKeys(ctx context.Context, inputs []*GhostInput, senders ...string) ([]*GhostKeys, error) {
	var (
		body interface{} = inputs
		resp []*GhostKeys
	)

	if len(senders) > 0 {
		body = map[string]interface{}{
			"keys":    inputs,
			"senders": senders,
		}
	}

	if err := c.Post(ctx, "/safe/keys", body, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func SafeCreateXinAddressGhostKeys(txVer uint8, addresses []*mixinnet.Address, outputIndex uint8) *GhostKeys {
	r := mixinnet.GenerateKey(rand.Reader)
	keys := &GhostKeys{
		Mask: r.Public(),
		Keys: make([]mixinnet.Key, len(addresses)),
	}

	for i, a := range addresses {
		k := mixinnet.DeriveGhostPublicKey(txVer, &r, &a.PublicViewKey, &a.PublicSpendKey, outputIndex)
		keys.Keys[i] = *k
	}

	return keys
}
