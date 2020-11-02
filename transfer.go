package mixin

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// TransferInput input for transfer/verify payment request
type TransferInput struct {
	AssetID    string          `json:"asset_id,omitempty"`
	OpponentID string          `json:"opponent_id,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	TraceID    string          `json:"trace_id,omitempty"`
	Memo       string          `json:"memo,omitempty"`

	// OpponentKey used for raw transaction
	OpponentKey string `json:"opponent_key,omitempty"`

	OpponentMultisig struct {
		Receivers []string `json:"receivers,omitempty"`
		Threshold uint8    `json:"threshold,omitempty"`
	} `json:"opponent_multisig,omitempty"`
}

func (c *Client) Transfer(ctx context.Context, input *TransferInput, pin string) (*Snapshot, error) {
	body := struct {
		*TransferInput
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

func (c *Client) ReadTransfer(ctx context.Context, traceID string) (*Snapshot, error) {
	uri := fmt.Sprintf("/transfers/trace/%s", traceID)

	var snapshot Snapshot
	if err := c.Get(ctx, uri, nil, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}
