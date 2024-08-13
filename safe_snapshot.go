package mixin

import (
	"context"
	"strconv"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

type (
	SafeSnapshotDeposit struct {
		DepositHash  string `json:"deposit_hash,omitempty"`
		DepositIndex uint64 `json:"deposit_index,omitempty"`
		Sender       string `json:"sender,omitempty"`
	}

	SafeSnapshot struct {
		SnapshotID      string               `json:"snapshot_id,omitempty"`
		RequestID       string               `json:"request_id,omitempty"`
		UserID          string               `json:"user_id,omitempty"`
		OpponentID      string               `json:"opponent_id,omitempty"`
		TransactionHash *mixinnet.Hash       `json:"transaction_hash,omitempty"`
		AssetID         string               `json:"asset_id,omitempty"`
		KernelAssetID   string               `json:"kernel_asset_id,omitempty"`
		Amount          decimal.Decimal      `json:"amount,omitempty"`
		Memo            string               `json:"memo,omitempty"`
		CreatedAt       time.Time            `json:"created_at"`
		Deposit         *SafeSnapshotDeposit `json:"deposit,omitempty"`
	}
)

func (c *Client) ReadSafeSnapshot(ctx context.Context, snapshotID string) (*SafeSnapshot, error) {
	var snapshot SafeSnapshot
	if err := c.Get(ctx, "/safe/snapshots/"+snapshotID, nil, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func ReadSafeSnapshot(ctx context.Context, accessToken, snapshotID string) (*SafeSnapshot, error) {
	return NewFromAccessToken(accessToken).ReadSafeSnapshot(ctx, snapshotID)
}

func (c *Client) ReadSafeSnapshots(ctx context.Context, assetID string, offset time.Time, order string, limit int) ([]*SafeSnapshot, error) {
	params := buildReadSafeSnapshotsParams("", assetID, offset, order, limit)

	var snapshots []*SafeSnapshot
	if err := c.Get(ctx, "/safe/snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func ReadSafeSnapshots(ctx context.Context, accessToken string, assetID string, offset time.Time, order string, limit int) ([]*SafeSnapshot, error) {
	return NewFromAccessToken(accessToken).ReadSafeSnapshots(ctx, assetID, offset, order, limit)
}

// list safe snapshots of dapp & sub wallets created by this dapp
func (c *Client) ReadSafeAppSnapshots(ctx context.Context, assetID string, offset time.Time, order string, limit int) ([]*SafeSnapshot, error) {
	params := buildReadSafeSnapshotsParams(c.ClientID, assetID, offset, order, limit)

	var snapshots []*SafeSnapshot
	if err := c.Get(ctx, "/safe/snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func buildReadSafeSnapshotsParams(appID string, assetID string, offset time.Time, order string, limit int) map[string]string {
	params := make(map[string]string)

	if appID != "" {
		params["app"] = appID
	}

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

func (c *Client) SafeNotifySnapshot(ctx context.Context, receiverID string, hash mixinnet.Hash, index uint8) error {
	const uri = "/safe/snapshots/notifications"
	return c.Post(ctx, uri, map[string]interface{}{
		"transaction_hash": hash.String(),
		"output_index":     index,
		"receiver_id":      receiverID,
	}, nil)
}
