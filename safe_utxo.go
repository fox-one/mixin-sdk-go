package mixin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

const (
	SafeUtxoStateUnspent = "unspent"
	SafeUtxoStateSigned  = "signed"
	SafeUtxoStateSpent   = "spent"
)

type (
	SafeUtxoState string

	SafeUtxo struct {
		OutputID           string          `json:"output_id,omitempty"`
		TransactionHash    Hash            `json:"transaction_hash,omitempty"`
		OutputIndex        uint64          `json:"output_index,omitempty"`
		Asset              Hash            `json:"asset,omitempty"`
		Amount             decimal.Decimal `json:"amount,omitempty"`
		Mask               Key             `json:"mask,omitempty"`
		Keys               []Key           `json:"keys,omitempty"`
		SendersHash        string          `json:"senders_hash,omitempty"`
		SendersThreshold   uint8           `json:"senders_threshold,omitempty"`
		Senders            []string        `json:"senders,omitempty"`
		ReceiversHash      Hash            `json:"receivers_hash,omitempty"`
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
	Members   []string
	Threshold uint8
	Offset    uint64
	Limit     int
	Order     string
	State     string
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
	params["members"] = HashMembers(opt.Members)
	params["threshold"] = fmt.Sprint(opt.Threshold)

	switch opt.Order {
	case "ASC", "DESC":
	default:
		opt.Order = "DESC"
	}
	params["order"] = opt.Order

	if opt.State != "" {
		params["state"] = opt.State
	}

	var utxos = []*SafeUtxo{}
	if err := c.Get(ctx, "/safe/outputs", params, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}
