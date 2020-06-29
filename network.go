package mixin

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type NetworkChain struct {
	ChainID              string          `json:"chain_id"`
	IconURL              string          `json:"icon_url"`
	Name                 string          `json:"name"`
	Type                 string          `json:"type"`
	WithdrawFee          decimal.Decimal `json:"withdrawal_fee"`
	WithdrawTimestamp    time.Time       `json:"withdrawal_timestamp"`
	WithdrawPendingCount int64           `json:"withdrawal_pending_count"`
	DepositBlockHeight   int64           `json:"deposit_block_height"`
	ExternalBlockHeight  int64           `json:"external_block_height"`
	ManagedBlockHeight   int64           `json:"managed_block_height"`
	IsSynchronized       bool            `json:"is_synchronized"`
}

type NetworkAsset struct {
	Amount  decimal.Decimal `json:"amount"`
	AssetID string          `json:"asset_id"`
	IconURL string          `json:"icon_url"`
	Symbol  string          `json:"symbol"`
}

// NetworkInfo mixin network info
type NetworkInfo struct {
	Assets         []*NetworkAsset `json:"assets"`
	Chains         []*NetworkChain `json:"chains"`
	AssetsCount    decimal.Decimal `json:"assets_count"`
	PeakThroughput decimal.Decimal `json:"peak_throughput"`
	SnapshotsCount decimal.Decimal `json:"snapshots_count"`
	Type           string          `json:"type"`
}

// ReadNetworkInfo read mixin network
func ReadNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	resp, err := Request(ctx).Get("/network")
	if err != nil {
		return nil, err
	}

	var info NetworkInfo
	if err := UnmarshalResponse(resp, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
