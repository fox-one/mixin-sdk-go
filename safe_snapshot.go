package mixin

import (
	"context"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type (
	SafeSnapshot struct {
		SnapshotID      string          `json:"snapshot_id,omitempty"`
		UserID          string          `json:"user_id,omitempty"`
		OpponentID      string          `json:"opponent_id,omitempty"`
		TransactionHash *Hash           `json:"transaction_hash,omitempty"`
		AssetID         string          `json:"asset_id,omitempty"`
		Amount          decimal.Decimal `json:"amount,omitempty"`
		Memo            string          `json:"memo,omitempty"`
		CreatedAt       time.Time       `json:"created_at"`
	}
)

func (c *Client) ReadSafeSnapshot(ctx context.Context, snapshotID string) (*SafeSnapshot, error) {
	var snapshot SafeSnapshot
	if err := c.Get(ctx, "/safe/snapshots/"+snapshotID, nil, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func ReadSafeSnapshot(ctx context.Context, accessToken string, assetID string, offset time.Time, order string, limit int) ([]*SafeSnapshot, error) {
	return NewFromAccessToken(accessToken).ReadSafeSnapshots(ctx, assetID, offset, order, limit)
}

func (c *Client) ReadSafeSnapshots(ctx context.Context, assetID string, offset time.Time, order string, limit int) ([]*SafeSnapshot, error) {
	params := buildReadSafeSnapshotsParams(assetID, offset, order, limit)

	var snapshots []*SafeSnapshot
	if err := c.Get(ctx, "/safe/snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func ReadSafeSnapshots(ctx context.Context, accessToken string, assetID string, offset time.Time, order string, limit int) ([]*SafeSnapshot, error) {
	return NewFromAccessToken(accessToken).ReadSafeSnapshots(ctx, assetID, offset, order, limit)
}

func buildReadSafeSnapshotsParams(assetID string, offset time.Time, order string, limit int) map[string]string {
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
