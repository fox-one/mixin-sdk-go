package mixin

import (
	"context"

	"github.com/shopspring/decimal"
)

// TransferInput input for transfer/verify payment request
type TransferInput struct {
	AssetID    string          `json:"asset_id,omitempty"`
	OpponentID string          `json:"opponent_id,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	TraceID    string          `json:"trace_id,omitempty"`
	Memo       string          `json:"memo,omitempty"`
}

func (c *Client) Transfer(ctx context.Context, input TransferInput, pin string) (*Snapshot, error) {
	body := struct {
		TransferInput
		Pin string
	}{
		TransferInput: input,
		Pin:           c.EncryptPin(pin),
	}

	var snapshot Snapshot
	if err := c.Post(ctx, "/transfers", body, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}
