package mixin

import (
	"context"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

type WithdrawInput struct {
	AddressID string          `json:"address_id,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	Memo      string          `json:"memo,omitempty"`
}

func (c *Client) Withdraw(ctx context.Context, input WithdrawInput, pin string) (*Snapshot, error) {
	var body interface{}
	if key, err := mixinnet.KeyFromString(pin); err == nil {
		body = struct {
			WithdrawInput
			Pin string `json:"pin_base64"`
		}{
			WithdrawInput: input,
			Pin: c.EncryptTipPin(
				key,
				TIPWithdrawalCreate,
				input.AddressID,
				input.Amount.String(),
				"0", // fee
				input.TraceID,
				input.Memo,
			),
		}
	} else {
		body = struct {
			WithdrawInput
			Pin string
		}{
			WithdrawInput: input,
			Pin:           c.EncryptPin(pin),
		}
	}

	var snapshot Snapshot
	if err := c.Post(ctx, "/withdrawals", body, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}
