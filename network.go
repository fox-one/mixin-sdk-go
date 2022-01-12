package mixin

import (
	"context"
	"fmt"
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

type Ticker struct {
	Type     string          `json:"type"`
	PriceUSD decimal.Decimal `json:"price_usd"`
	PriceBTC decimal.Decimal `json:"price_btc"`
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

// ReadNetworkAsset read mixin network asset by asset id
func ReadNetworkAsset(ctx context.Context, assetID string) (*Asset, error) {
	uri := fmt.Sprintf("/network/assets/%s", assetID)

	resp, err := Request(ctx).Get(uri)
	if err != nil {
		return nil, err
	}

	var asset Asset
	if err := UnmarshalResponse(resp, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

// ReadTopNetworkAssets read top network assets
func ReadTopNetworkAssets(ctx context.Context) ([]*Asset, error) {
	resp, err := Request(ctx).Get("/network/assets/top")
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if err := UnmarshalResponse(resp, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

// ReadNetworkAssetsBySymbol read mixin network assets by symbol
func ReadNetworkAssetsBySymbol(ctx context.Context, symbol string) ([]*Asset, error) {
	uri := fmt.Sprintf("/network/assets/search/%s", symbol)

	resp, err := Request(ctx).Get(uri)
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if err := UnmarshalResponse(resp, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

// ReadTicker read mixin ticker of asset with offset
func ReadTicker(ctx context.Context, assetID string, offset time.Time) (*Ticker, error) {
	params := map[string]string{
		"asset": assetID,
	}
	if !offset.IsZero() {
		params["offset"] = offset.Format(time.RFC3339Nano)
	}
	resp, err := Request(ctx).SetQueryParams(params).Get("/network/ticker")
	if err != nil {
		return nil, err
	}

	var ticker Ticker
	if err := UnmarshalResponse(resp, &ticker); err != nil {
		return nil, err
	}
	return &ticker, nil
}
