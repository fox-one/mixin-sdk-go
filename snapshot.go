package mixin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	if sj.AssetID == "" && sj.Asset != nil {
		sj.AssetID = sj.Asset.AssetID
	}

	*s = (Snapshot)(sj.snapshotJSON)
	return nil
}

// ReadSnapshots return a list of snapshots
// order must be `ASC` or `DESC`
func (c *Client) ReadSnapshots(ctx context.Context, assetID string, offset time.Time, order string, limit int) ([]*Snapshot, error) {
	var snapshots []*Snapshot
	params := buildReadSnapshotsParams(assetID, offset, order, limit)
	if err := c.Get(ctx, "/snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

// ReadSnapshots by accessToken, scope SNAPSHOTS:READ required
func ReadSnapshots(ctx context.Context, accessToken string, assetID string, offset time.Time, order string, limit int) ([]*Snapshot, error) {
	return NewFromAccessToken(accessToken).ReadSnapshots(ctx, assetID, offset, order, limit)
}

func (c *Client) ReadNetworkSnapshots(ctx context.Context, assetID string, offset time.Time, order string, limit int) ([]*Snapshot, error) {
	var snapshots []*Snapshot
	params := buildReadSnapshotsParams(assetID, offset, order, limit)
	if err := c.Get(ctx, "/network/snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func (c *Client) ReadSnapshot(ctx context.Context, snapshotID string) (*Snapshot, error) {
	uri := fmt.Sprintf("/snapshots/%s", snapshotID)

	var snapshot Snapshot
	if err := c.Get(ctx, uri, nil, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// ReadSnapshot by accessToken, scope SNAPSHOTS:READ required
func ReadSnapshot(ctx context.Context, accessToken, snapshotID string) (*Snapshot, error) {
	return NewFromAccessToken(accessToken).ReadSnapshot(ctx, snapshotID)
}

func (c *Client) ReadSnapshotByTraceID(ctx context.Context, traceID string) (*Snapshot, error) {
	uri := fmt.Sprintf("/snapshots/trace/%s", traceID)

	var snapshot Snapshot
	if err := c.Get(ctx, uri, nil, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// ReadSnapshotByTraceID by accessToken, scope SNAPSHOTS:READ required
func ReadSnapshotByTraceID(ctx context.Context, accessToken, traceID string) (*Snapshot, error) {
	return NewFromAccessToken(accessToken).ReadSnapshotByTraceID(ctx, traceID)
}

func (c *Client) ReadNetworkSnapshot(ctx context.Context, snapshotID string) (*Snapshot, error) {
	uri := fmt.Sprintf("/network/snapshots/%s", snapshotID)

	var snapshot Snapshot
	if err := c.Get(ctx, uri, nil, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func buildReadSnapshotsParams(assetID string, offset time.Time, order string, limit int) map[string]string {
	params := make(map[string]string)

	if assetID != "" {
		params["asset"] = assetID
	}

	if !offset.IsZero() {
		params["offset"] = offset.UTC().Format(time.RFC3339Nano)
	}

	switch order {
	case "ASC", "DESC":
	default:
		order = "DESC"
	}

	params["order"] = order

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	return params
}
