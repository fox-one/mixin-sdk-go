package mixin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

const (
	SafeUtxoStateUnspent SafeUtxoState = "unspent"
	SafeUtxoStateSigned  SafeUtxoState = "signed"
	SafeUtxoStateSpent   SafeUtxoState = "spent"
)

type (
	SafeUtxoState string

	SafeUtxo struct {
		OutputID           string          `json:"output_id,omitempty"`
		RequestID          string          `json:"request_id,omitempty"`
		TransactionHash    mixinnet.Hash   `json:"transaction_hash,omitempty"`
		OutputIndex        uint8           `json:"output_index,omitempty"`
		KernelAssetID      mixinnet.Hash   `json:"kernel_asset_id,omitempty"`
		AssetID            string          `json:"asset_id,omitempty"`
		Amount             decimal.Decimal `json:"amount,omitempty"`
		Mask               mixinnet.Key    `json:"mask,omitempty"`
		Keys               []mixinnet.Key  `json:"keys,omitempty"`
		InscriptionHash    mixinnet.Hash   `json:"inscription_hash,omitempty"`
		SendersHash        string          `json:"senders_hash,omitempty"`
		SendersThreshold   uint8           `json:"senders_threshold,omitempty"`
		Senders            []string        `json:"senders,omitempty"`
		ReceiversHash      mixinnet.Hash   `json:"receivers_hash,omitempty"`
		ReceiversThreshold uint8           `json:"receivers_threshold,omitempty"`
		Receivers          []string        `json:"receivers,omitempty"`
		Extra              string          `json:"extra,omitempty"`
		State              SafeUtxoState   `json:"state,omitempty"`
		Sequence           uint64          `json:"sequence,omitempty"`
		CreatedAt          time.Time       `json:"created_at,omitempty"`
		UpdatedAt          time.Time       `json:"updated_at,omitempty"`
		Signers            []string        `json:"signers,omitempty"`
		SignedBy           string          `json:"signed_by,omitempty"`
		SignedAt           *time.Time      `json:"signed_at,omitempty"`
		SpentAt            *time.Time      `json:"spent_at,omitempty"`
	}
)

type SafeListUtxoOption struct {
	Members           []string
	Threshold         uint8
	Offset            uint64
	Asset             string
	Limit             int
	Order             string
	State             SafeUtxoState
	IncludeSubWallets bool
}

func (c *Client) SafeListUtxos(ctx context.Context, opt SafeListUtxoOption) ([]*SafeUtxo, error) {
	params := make(map[string]string)
	if opt.Offset > 0 {
		params["offset"] = fmt.Sprint(opt.Offset)
	}

	if opt.Limit > 0 {
		params["limit"] = strconv.Itoa(opt.Limit)
	}

	if len(opt.Members) == 0 {
		opt.Members = []string{c.ClientID}
	}

	if opt.Threshold < 1 {
		opt.Threshold = 1
	}
	if int(opt.Threshold) > len(opt.Members) {
		return nil, errors.New("invalid members")
	}
	params["members"] = mixinnet.HashMembers(opt.Members)
	params["threshold"] = fmt.Sprint(opt.Threshold)

	if opt.IncludeSubWallets {
		params["app"] = c.ClientID
	}

	switch opt.Order {
	case "ASC", "DESC":
	default:
		opt.Order = "DESC"
	}
	params["order"] = opt.Order

	if opt.State != "" {
		params["state"] = string(opt.State)
	}

	if opt.Asset != "" {
		params["asset"] = opt.Asset
	}

	var utxos []*SafeUtxo
	if err := c.Get(ctx, "/safe/outputs", params, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}

func (c *Client) SafeReadUtxo(ctx context.Context, id string) (*SafeUtxo, error) {
	uri := fmt.Sprintf("/safe/outputs/%s", id)

	var utxo SafeUtxo
	if err := c.Get(ctx, uri, nil, &utxo); err != nil {
		return nil, err
	}

	return &utxo, nil
}

func (c *Client) SafeReadUtxoByHash(ctx context.Context, hash mixinnet.Hash, index uint8) (*SafeUtxo, error) {
	id := fmt.Sprintf("%s:%d", hash.String(), index)
	return c.SafeReadUtxo(ctx, id)
}
