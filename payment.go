package mixin

import (
	"context"

	"github.com/shopspring/decimal"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
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

type Payment struct {
	Recipient *User      `json:"recipient,omitempty"`
	Asset     *SafeAsset `json:"asset,omitempty"`
	AssetID   string     `json:"asset_id,omitempty"`
	Amount    string     `json:"amount,omitempty"`
	TraceID   string     `json:"trace_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	Memo      string     `json:"memo,omitempty"`
	Receivers []string   `json:"receivers,omitempty"`
	Threshold uint8      `json:"threshold,omitempty"`
	CodeID    string     `json:"code_id,omitempty"`
}

func (c *Client) VerifyPayment(ctx context.Context, input TransferInput) (*Payment, error) {
	var resp Payment
	if err := c.Post(ctx, "/payments", input, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
