package mixin

import (
	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

type (
	SafeTransactionRequestInput struct {
		RequestID      string `json:"request_id"`
		RawTransaction string `json:"raw"`
	}

	SafeTransactionReceiver struct {
		Members    []string      `json:"members,omitempty"`
		MemberHash mixinnet.Hash `json:"members_hash,omitempty"`
		Threshold  uint8         `json:"threshold,omitempty"`
	}

	SafeTransactionRequest struct {
		RequestID        string                     `json:"request_id,omitempty"`
		TransactionHash  string                     `json:"transaction_hash,omitempty"`
		UserID           string                     `json:"user_id,omitempty"`
		KernelAssetID    mixinnet.Hash              `json:"kernel_asset_id,omitempty"`
		AssetID          mixinnet.Hash              `json:"asset_id,omitempty"`
		Amount           decimal.Decimal            `json:"amount,omitempty"`
		CreatedAt        time.Time                  `json:"created_at,omitempty"`
		UpdatedAt        time.Time                  `json:"updated_at,omitempty"`
		Extra            string                     `json:"extra,omitempty"`
		Receivers        []*SafeTransactionReceiver `json:"receivers,omitempty"`
		Senders          []string                   `json:"senders,omitempty"`
		SendersHash      string                     `json:"senders_hash,omitempty"`
		SendersThreshold uint8                      `json:"senders_threshold,omitempty"`
		Signers          []string                   `json:"signers,omitempty"`
		SnapshotHash     string                     `json:"snapshot_hash,omitempty"`
		SnapshotAt       *time.Time                 `json:"snapshot_at,omitempty"`
		State            SafeUtxoState              `json:"state,omitempty"`
		RawTransaction   string                     `json:"raw_transaction"`
		Views            []mixinnet.Key             `json:"views,omitempty"`
		RevokedBy        string                     `json:"revoked_by"`

		// TODO delete when asset_id is on
		Asset mixinnet.Hash `json:"asset,omitempty"`
	}
)

func (c *Client) SafeCreateTransactionRequests(ctx context.Context, inputs []*SafeTransactionRequestInput) ([]*SafeTransactionRequest, error) {
	var resp []*SafeTransactionRequest
	if err := c.Post(ctx, "/safe/transaction/requests", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SafeCreateTransactionRequest(ctx context.Context, input *SafeTransactionRequestInput) (*SafeTransactionRequest, error) {
	requests, err := c.SafeCreateTransactionRequests(ctx, []*SafeTransactionRequestInput{input})
	if err != nil {
		return nil, err
	}

	return requests[0], nil
}

func (c *Client) SafeReadTransactionRequest(ctx context.Context, idOrHash string) (*SafeTransactionRequest, error) {
	var resp SafeTransactionRequest
	if err := c.Get(ctx, "/safe/transactions/"+idOrHash, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) SafeSubmitTransactionRequests(ctx context.Context, inputs []*SafeTransactionRequestInput) ([]*SafeTransactionRequest, error) {
	var resp []*SafeTransactionRequest
	if err := c.Post(ctx, "/safe/transactions", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SafeSubmitTransactionRequest(ctx context.Context, input *SafeTransactionRequestInput) (*SafeTransactionRequest, error) {
	requests, err := c.SafeSubmitTransactionRequests(ctx, []*SafeTransactionRequestInput{input})
	if err != nil {
		return nil, err
	}

	return requests[0], nil
}
