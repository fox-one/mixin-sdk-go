package mixin

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MixinNetwork/mixin/common"
	"github.com/shopspring/decimal"
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

// GhostKeys transaction ghost keys
type GhostKeys struct {
	Mask string   `json:"mask"`
	Keys []string `json:"keys"`
}

func (g GhostKeys) DumpOutput(threshold int, amount decimal.Decimal) *Output {
	return &Output{
		Mask:   g.Mask,
		Keys:   g.Keys,
		Amount: amount.Truncate(8).String(),
		Script: common.NewThresholdScript(uint8(threshold)).String(),
	}
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

func (c *Client) ReadGhostKeys(ctx context.Context, receivers []string, index int) (*GhostKeys, error) {
	data, err := json.Marshal(map[string]interface{}{
		"receivers": receivers,
		"index":     index,
	})
	if err != nil {
		return nil, err
	}

	var resp GhostKeys
	if err := c.Post(ctx, "/outputs", data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
