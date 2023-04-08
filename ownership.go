package mixin

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) TransferOwnership(ctx context.Context, newOwner, pin string) error {
	key, err := KeyFromString(pin)
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

	var resp json.RawMessage
	uri := fmt.Sprintf("/apps/%s/transfer", c.ClientID)
	if err := c.Post(ctx, uri, body, &resp); err != nil {
		return err
	}

	{
		bts, _ := json.MarshalIndent(resp, "", "    ")
		fmt.Println(string(bts))
	}

	return nil
}
