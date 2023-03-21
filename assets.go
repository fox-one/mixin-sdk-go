package mixin

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

type Asset struct {
	AssetID        string          `json:"asset_id"`
	ChainID        string          `json:"chain_id"`
	AssetKey       string          `json:"asset_key,omitempty"`
	MixinID        string          `json:"mixin_id,omitempty"`
	Symbol         string          `json:"symbol,omitempty"`
	Name           string          `json:"name,omitempty"`
	IconURL        string          `json:"icon_url,omitempty"`
	PriceBTC       decimal.Decimal `json:"price_btc,omitempty"`
	ChangeBTC      decimal.Decimal `json:"change_btc,omitempty"`
	PriceUSD       decimal.Decimal `json:"price_usd,omitempty"`
	ChangeUsd      decimal.Decimal `json:"change_usd,omitempty"`
	Balance        decimal.Decimal `json:"balance,omitempty"`
	Destination    string          `json:"destination,omitempty"`
	Tag            string          `json:"tag,omitempty"`
	Confirmations  int             `json:"confirmations,omitempty"`
	Capitalization float64         `json:"capitalization,omitempty"`
	DepositEntries []DepositEntry  `json:"deposit_entries"`
}

type DepositEntry struct {
	Destination string   `json:"destination"`
	Tag         string   `json:"tag"`
	Properties  []string `json:"properties"`
}

func (c *Client) ReadAsset(ctx context.Context, assetID string) (*Asset, error) {
	uri := fmt.Sprintf("/assets/%s", assetID)

	var asset Asset
	if err := c.Get(ctx, uri, nil, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

func ReadAsset(ctx context.Context, accessToken, assetID string) (*Asset, error) {
	return NewFromAccessToken(accessToken).ReadAsset(ctx, assetID)
}

func (c *Client) ReadAssets(ctx context.Context) ([]*Asset, error) {
	var assets []*Asset
	if err := c.Get(ctx, "/assets", nil, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

func ReadAssets(ctx context.Context, accessToken string) ([]*Asset, error) {
	return NewFromAccessToken(accessToken).ReadAssets(ctx)
}

func (c *Client) ReadAssetFee(ctx context.Context, assetID string) (decimal.Decimal, error) {
	uri := fmt.Sprintf("/assets/%s/fee", assetID)

	var body struct {
		Amount decimal.Decimal `json:"amount,omitempty"`
	}
	if err := c.Get(ctx, uri, nil, &body); err != nil {
		return decimal.Zero, nil
	}

	return body.Amount, nil
}
