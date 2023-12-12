package mixin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

type (
	// RawTransaction raw transaction
	RawTransaction struct {
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
)

func (c *Client) Transaction(ctx context.Context, in *TransferInput, pin string) (*RawTransaction, error) {
	paras := map[string]interface{}{
		"asset_id": in.AssetID,
		"amount":   in.Amount,
		"trace_id": in.TraceID,
		"memo":     in.Memo,
	}

	if key, err := mixinnet.KeyFromString(pin); err == nil {
		paras["pin_base64"] = c.EncryptTipPin(
			key,
			TIPRawTransactionCreate,
			in.AssetID,
			in.OpponentKey,
			strings.Join(in.OpponentMultisig.Receivers, ""),
			fmt.Sprint(in.OpponentMultisig.Threshold),
			in.Amount.String(),
			in.TraceID,
			in.Memo,
		)
	} else {
		paras["pin"] = c.EncryptPin(pin)
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
