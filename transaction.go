package mixin

import (
	"context"
	"time"
)

// RawTransaction raw transaction
type RawTransaction struct {
	Type            string    `json:"type"`
	SnapshotID      string    `json:"snapshot"`
	OpponentKey     string    `json:"opponent_key"`
	AssetID         string    `json:"asset_id"`
	Amount          string    `json:"amount"`
	TraceID         string    `json:"trace_id"`
	Memo            string    `json:"memo"`
	State           string    `json:"state"`
	CreatedAt       time.Time `json:"created_at"`
	TransactionHash string    `json:"transaction_hash"`
	SnapshotHash    string    `json:"snapshot_hash"`
	SnapshotAt      time.Time `json:"snapshot_at"`
}

// TransactionOutput transaction output
type TransactionOutput struct {
	Mask string   `json:"mask"`
	Keys []string `json:"keys"`
}

func (c *Client) Transaction(ctx context.Context, in *TransferInput, pin string) (*RawTransaction, error) {
	paras := map[string]interface{}{
		"asset_id":     in.AssetID,
		"opponent_key": in.OpponentKey,
		"amount":       in.Amount,
		"trace_id":     in.TraceID,
		"memo":         in.Memo,
		"pin":          c.EncryptPin(pin),
	}

	var resp RawTransaction
	if err := c.Post(ctx, "/transactions", paras, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) MakeTransactionOutput(ctx context.Context, ids ...string) (*TransactionOutput, error) {
	if len(ids) == 0 {
		ids = []string{c.ClientID}
	}

	var resp TransactionOutput
	if err := c.Post(ctx, "/outputs", ids, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
