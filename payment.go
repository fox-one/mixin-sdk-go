package mixin

import (
	"context"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
)

type Payment struct {
	Recipient *User  `json:"recipient,omitempty"`
	Asset     *Asset `json:"asset,omitempty"`
	Amount    string `json:"amount,omitempty"`
	Status    string `json:"status,omitempty"`
}

func (c *Client) VerifyPayment(ctx context.Context, input TransferInput) (*Payment, error) {
	var resp Payment
	if err := c.Post(ctx, "/payments", input, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
