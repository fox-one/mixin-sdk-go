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
		"pin": c.EncryptPin(newPin),
	}

	if pin != "" {
		body["old_pin"] = c.EncryptPin(pin)
	}

	return c.Post(ctx, "/pin/update", body, nil)
}
