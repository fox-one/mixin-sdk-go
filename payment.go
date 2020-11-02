package mixin

import (
	"context"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
)

type Payment struct {
	Recipient *User    `json:"recipient,omitempty"`
	Asset     *Asset   `json:"asset,omitempty"`
	AssetID   string   `json:"asset_id,omitempty"`
	Amount    string   `json:"amount,omitempty"`
	TraceID   string   `json:"trace_id,omitempty"`
	Status    string   `json:"status,omitempty"`
	Memo      string   `json:"memo,omitempty"`
	Receivers []string `json:"receivers,omitempty"`
	Threshold uint8    `json:"threshold,omitempty"`
	CodeID    string   `json:"code_id,omitempty"`
}

func (c *Client) VerifyPayment(ctx context.Context, input TransferInput) (*Payment, error) {
	var resp Payment
	if err := c.Post(ctx, "/payments", input, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
