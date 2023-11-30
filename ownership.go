package mixin

import (
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

func (c *Client) TransferOwnership(ctx context.Context, newOwner, pin string) error {
	key, err := mixinnet.KeyFromString(pin)
	if err != nil {
		return err
	}
	var body = struct {
		UserID    string `json:"user_id"`
		PinBase64 string `json:"pin_base64"`
	}{
		UserID: newOwner,
		PinBase64: c.EncryptTipPin(
			key,
			TIPAppOwnershipTransfer,
			newOwner,
		),
	}

	uri := fmt.Sprintf("/apps/%s/transfer", c.ClientID)
	if err := c.Post(ctx, uri, body, nil); err != nil {
		return err
	}

	return nil
}
