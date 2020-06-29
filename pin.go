package mixin

import (
	"context"
)

func (c *Client) VerifyPin(ctx context.Context, pin string) error {
	body := map[string]interface{}{
		"pin": c.EncryptPin(pin),
	}

	return c.Post(ctx, "/pin/verify", body, nil)
}

func (c *Client) ModifyPin(ctx context.Context, pin, newPin string) error {
	body := map[string]interface{}{
		"old_pin": c.EncryptPin(pin),
		"pin":     c.EncryptPin(newPin),
	}

	return c.Post(ctx, "/pin/update", body, nil)
}
