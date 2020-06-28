package mixin

import (
	"context"

	"github.com/shopspring/decimal"
)

type WithdrawInput struct {
	AddressID string          `json:"address_id,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	Memo      string          `json:"memo,omitempty"`
}

func (c *Client) Withdraw(ctx context.Context, input WithdrawInput, pin string) (*Snapshot, error) {
	body := struct {
		WithdrawInput
		Pin string
	}{
		WithdrawInput: input,
		Pin:           c.EncryptPin(pin),
	}

	var snapshot Snapshot
	if err := c.Post(ctx, "/withdrawals", body, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}
