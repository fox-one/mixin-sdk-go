package mixin

import (
	"context"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

func (c *Client) ParseTipPin(ctx context.Context, pin string) (mixinnet.Key, error) {
	if c.publicTipKey == "" {
		if _, err := c.UserMe(ctx); err != nil {
			return mixinnet.Key{}, err
		}
	}

	if key, err := mixinnet.KeyFromSeed(pin); err == nil && key.Public().String() == c.publicTipKey {
		return key, nil
	}

	if key, err := mixinnet.KeyFromString(pin); err == nil && key.Public().String() == c.publicTipKey {
		return key, nil
	}

	return mixinnet.Key{}, createError(403, PinIncorrect, "pin incorrect")
}

func (c *Client) ParseSpendKey(ctx context.Context, spend string) (mixinnet.Key, error) {
	if c.publicSpendKey == "" {
		if _, err := c.UserMe(ctx); err != nil {
			return mixinnet.Key{}, err
		}
	}

	if key, err := mixinnet.KeyFromSeed(spend); err == nil && key.Public().String() == c.publicSpendKey {
		return key, nil
	}

	if key, err := mixinnet.KeyFromString(spend); err == nil && key.Public().String() == c.publicSpendKey {
		return key, nil
	}

	return mixinnet.Key{}, createError(403, InvalidSpendKey, "spend key incorrect")
}
