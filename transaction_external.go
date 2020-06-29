package mixin

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type ExternalTransaction struct {
	TransactionID   string          `json:"transaction_id"`
	CreatedAt       time.Time       `json:"created_at"`
	TransactionHash string          `json:"transaction_hash"`
	Sender          string          `json:"sender"`
	ChainId         string          `json:"chain_id"`
	AssetId         string          `json:"asset_id"`
	Amount          decimal.Decimal `json:"amount"`
	Destination     string          `json:"destination"`
	Tag             string          `json:"tag"`
	Confirmations   int64           `json:"confirmations"`
	Threshold       int64           `json:"threshold"`
}

func ReadExternalTransactions(ctx context.Context, assetID, destination, tag string) ([]*ExternalTransaction, error) {
	params := make(map[string]string)
	if destination != "" {
		params["destination"] = destination
	}
	if tag != "" {
		params["tag"] = tag
	}
	if assetID != "" {
		params["asset"] = assetID
	}

	resp, err := Request(ctx).SetQueryParams(params).Get("/external/transactions")
	if err != nil {
		return nil, err
	}

	var transactions []*ExternalTransaction
	if err := UnmarshalResponse(resp, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
