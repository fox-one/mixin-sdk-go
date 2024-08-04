package mixin

import (
	"context"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

type (
	SafeMultisigRequest struct {
		RequestID        string          `json:"request_id,omitempty"`
		TransactionHash  string          `json:"transaction_hash,omitempty"`
		AssetID          string          `json:"asset_id,omitempty"`
		KernelAssetID    mixinnet.Hash   `json:"kernel_asset_id,omitempty"`
		Amount           decimal.Decimal `json:"amount,omitempty"`
		SendersHash      string          `json:"senders_hash,omitempty"`
		SendersThreshold uint8           `json:"senders_threshold,omitempty"`
		Senders          []string        `json:"senders,omitempty"`
		Signers          []string        `json:"signers,omitempty"`
		Extra            string          `json:"extra,omitempty"`
		RawTransaction   string          `json:"raw_transaction"`
		CreatedAt        time.Time       `json:"created_at,omitempty"`
		UpdatedAt        time.Time       `json:"updated_at,omitempty"`
		Views            []mixinnet.Key  `json:"views,omitempty"`
		RevokedBy        string          `json:"revoked_by"`
	}
)

func (c *Client) SafeCreateMultisigRequests(ctx context.Context, inputs []*SafeTransactionRequestInput) ([]*SafeMultisigRequest, error) {
	var resp []*SafeMultisigRequest
	if err := c.Post(ctx, "/safe/multisigs", inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) SafeReadMultisigRequests(ctx context.Context, idOrHash string) (*SafeMultisigRequest, error) {
	var resp SafeMultisigRequest
	if err := c.Get(ctx, "/safe/multisigs/"+idOrHash, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) SafeCreateMultisigRequest(ctx context.Context, input *SafeTransactionRequestInput) (*SafeMultisigRequest, error) {
	requests, err := c.SafeCreateMultisigRequests(ctx, []*SafeTransactionRequestInput{input})
	if err != nil {
		return nil, err
	}

	return requests[0], nil
}

func (c *Client) SafeSignMultisigRequest(ctx context.Context, input *SafeTransactionRequestInput) (*SafeMultisigRequest, error) {
	var resp SafeMultisigRequest
	uri := fmt.Sprintf("/safe/multisigs/%s/sign", input.RequestID)
	if err := c.Post(ctx, uri, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) SafeUnlockMultisigRequest(ctx context.Context, requestID string) (*SafeMultisigRequest, error) {
	var resp SafeMultisigRequest
	uri := fmt.Sprintf("/safe/multisigs/%s/unlock", requestID)
	if err := c.Post(ctx, uri, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
