package mixin

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// RawTransaction raw transaction
type RawTransaction struct {
	Type            string    `json:"type"`
	SnapshotID      string    `json:"snapshot_id,omitempty"`
	OpponentKey     string    `json:"opponent_key,omitempty"`
	AssetID         string    `json:"asset_id"`
	Amount          string    `json:"amount"`
	TraceID         string    `json:"trace_id"`
	Memo            string    `json:"memo"`
	State           string    `json:"state"`
	CreatedAt       time.Time `json:"created_at"`
	TransactionHash string    `json:"transaction_hash,omitempty"`
	SnapshotHash    string    `json:"snapshot_hash,omitempty"`
	SnapshotAt      time.Time `json:"snapshot_at"`
}

// GhostKeys transaction ghost keys
type GhostKeys struct {
	Mask Key   `json:"mask"`
	Keys []Key `json:"keys"`
}

func (g GhostKeys) DumpOutput(threshold uint8, amount decimal.Decimal) *Output {
	return &Output{
		Mask:   g.Mask,
		Keys:   g.Keys,
		Amount: NewIntegerFromDecimal(amount),
		Script: NewThresholdScript(threshold),
	}
}

func (c *Client) Transaction(ctx context.Context, in *TransferInput, pin string) (*RawTransaction, error) {
	paras := map[string]interface{}{
		"asset_id": in.AssetID,
		"amount":   in.Amount,
		"trace_id": in.TraceID,
		"memo":     in.Memo,
		"pin":      c.EncryptPin(pin),
	}

	if in.OpponentKey != "" {
		paras["opponent_key"] = in.OpponentKey
	} else {
		paras["opponent_multisig"] = map[string]interface{}{
			"receivers": in.OpponentMultisig.Receivers,
			"threshold": in.OpponentMultisig.Threshold,
		}
	}

	var resp RawTransaction
	if err := c.Post(ctx, "/transactions", paras, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type GhostInput struct {
	Receivers []string `json:"receivers"`
	Index     int      `json:"index"`
	Hint      string   `json:"hint"`
}

func (c *Client) ReadGhostKeys(ctx context.Context, receivers []string, index int) (*GhostKeys, error) {
	input := &GhostInput{
		Receivers: receivers,
		Index:     index,
		Hint:      "",
	}

	var resp GhostKeys
	if err := c.Post(ctx, "/outputs", input, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) BatchReadGhostKeys(ctx context.Context, inputs []*GhostInput) ([]*GhostKeys, error) {
	var resp []*GhostKeys
	if err := c.Post(ctx, "/outputs", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
