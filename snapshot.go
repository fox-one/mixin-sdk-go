package mixin

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type Snapshot struct {
	SnapshotID      string          `json:"snapshot_id"`
	CreatedAt       time.Time       `json:"created_at,omitempty"`
	TraceID         string          `json:"trace_id,omitempty"`
	UserID          string          `json:"user_id,omitempty"`
	AssetID         string          `json:"asset_id,omitempty"`
	ChainID         string          `json:"chain_id,omitempty"`
	OpponentID      string          `json:"opponent_id,omitempty"`
	Source          string          `json:"source,omitempty"`
	Amount          decimal.Decimal `json:"amount,omitempty"`
	OpeningBalance  decimal.Decimal `json:"opening_balance,omitempty"`
	ClosingBalance  decimal.Decimal `json:"closing_balance,omitempty"`
	Memo            string          `json:"memo,omitempty"`
	Type            string          `json:"type,omitempty"`
	Sender          string          `json:"sender,omitempty"`
	Receiver        string          `json:"receiver,omitempty"`
	TransactionHash string          `json:"transaction_hash,omitempty"`

	Asset *Asset `json:"asset,omitempty"`
}

type (
	snapshotJSON         Snapshot
	snapshotJSONWithData struct {
		snapshotJSON
		Data string `json:"data,omitempty"`
	}
)

func (s *Snapshot) UnmarshalJSON(b []byte) error {
	var sj snapshotJSONWithData
	if err := json.Unmarshal(b, &sj); err != nil {
		return err
	}

	if sj.Memo == "" {
		sj.Memo = sj.Data
	}

	*s = (Snapshot)(sj.snapshotJSON)
	return nil
}
