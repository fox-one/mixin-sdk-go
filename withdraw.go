package mixin

import (
	"context"
	"crypto/sha256"
	"fmt"

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
	if len(pin) == 6 {
		body = struct {
			WithdrawInput
			Pin string
		}{
			WithdrawInput: input,
			Pin:           c.EncryptPin(pin),
		}
	} else {
		key, err := KeyFromString(pin)
		if err != nil {
			return nil, err
		}

		addr, err := c.ReadAddress(ctx, input.AddressID)
		if err != nil {
			return nil, err
		}

		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s%s%s%s%s%s",
			TIPWithdrawalCreate,
			input.AddressID,
			input.Amount.String(), addr.Fee.String(),
			input.TraceID, input.Memo)))
		tipBody := hash.Sum(nil)
		pin = key.Sign(tipBody).String()

		body = struct {
			WithdrawInput
			Fee string `json:"fee"`
			Pin string `json:"pin_base64"`
		}{
			WithdrawInput: input,
			Fee:           addr.Fee.String(),
			Pin:           c.EncryptPin(pin),
		}
	}

	var snapshot Snapshot
	if err := c.Post(ctx, "/withdrawals", body, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}
