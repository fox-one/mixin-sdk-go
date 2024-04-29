package mixin

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type (
	SafeDepositEntry struct {
		EntryID     string   `json:"entry_id,omitempty"`
		Members     []string `json:"members,omitempty"`
		Threshold   int      `json:"threshold,omitempty"`
		ChainID     string   `json:"chain_id,omitempty"`
		Destination string   `json:"destination,omitempty"`
		Tag         string   `json:"tag,omitempty"`
		IsPrimary   bool     `json:"is_primary,omitempty"`
		Signature   string   `json:"signature,omitempty"`
	}

	SafeDeposit struct {
		DepositID       string          `json:"deposit_id,omitempty"`
		Destination     string          `json:"destination,omitempty"`
		Tag             string          `json:"tag,omitempty"`
		ChainID         string          `json:"chain_id,omitempty"`
		AssetID         string          `json:"asset_id,omitempty"`
		KernelAssetID   string          `json:"kernel_asset_id,omitempty"`
		AssetKey        string          `json:"chain_key,omitempty"`
		Amount          decimal.Decimal `json:"amount,omitempty"`
		TransactionHash string          `json:"transaction_hash,omitempty"`
		OutputIndex     uint64          `json:"output_index,omitempty"`
		BlockHash       string          `json:"block_hash,omitempty"`
		Confirmations   uint64          `json:"confirmations,omitempty"`
		Threshold       uint64          `json:"threshold,omitempty"`
		CreatedAt       time.Time       `json:"created_at,omitempty"`
		UpdatedAt       time.Time       `json:"updated_at,omitempty"`
	}
)

func (c *Client) SafeCreateDepositEntries(ctx context.Context, receivers []string, threshold int, chain string) ([]*SafeDepositEntry, error) {
	if len(receivers) == 0 {
		receivers = []string{c.ClientID}
	}
	if threshold < 1 {
		threshold = 1
	} else if threshold > len(receivers) {
		return nil, fmt.Errorf("invalid threshold %d, expect [1 %d)", threshold, len(receivers))
	}
	paras := map[string]interface{}{
		"members":   receivers,
		"threshold": threshold,
		"chain_id":  chain,
	}
	var entries []*SafeDepositEntry
	if err := c.Post(ctx, "/safe/deposit/entries", paras, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *Client) SafeListDeposits(ctx context.Context, entry *SafeDepositEntry, asset string, offset time.Time, limit int) ([]*SafeDeposit, error) {
	paras := map[string]string{
		"chain_id": entry.ChainID,
		"limit":    fmt.Sprint(limit),
	}

	if !offset.IsZero() {
		paras["offset"] = offset.Format(time.RFC3339Nano)
	}

	if entry.Destination != "" {
		paras["destination"] = entry.Destination

		if entry.Tag != "" {
			paras["tag"] = entry.Tag
		}
	}
	if asset != "" {
		paras["asset"] = asset
	}

	var deposits []*SafeDeposit
	if err := c.Get(ctx, "/safe/deposits", paras, &deposits); err != nil {
		return nil, err
	}

	return deposits, nil
}
