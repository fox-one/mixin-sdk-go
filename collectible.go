package mixin

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

const (
	// CollectibleOutputState
	CollectibleOutputStateUnspent = "unspent"
	CollectibleOutputStateSigned  = "signed"
	CollectibleOutputStateSpent   = "spent"

	// CollectibleRequestAction
	CollectibleRequestActionSign   = "sign"
	CollectibleRequestActionUnlock = "unlock"

	// CollectibleRequestState
	CollectibleRequestStateInitial = "initial"
	CollectibleRequestStateSigned  = "signed"
)

type CollectibleOutput struct {
	Type               string          `json:"type,omitempty"`
	CreatedAt          time.Time       `json:"created_at,omitempty"`
	UpdatedAt          time.Time       `json:"updated_at,omitempty"`
	UserID             string          `json:"user_id,omitempty"`
	OutputID           string          `json:"output_id,omitempty"`
	TokenID            string          `json:"token_id,omitempty"`
	Extra              string          `json:"extra,omitempty"`
	TransactionHash    Hash            `json:"transaction_hash,omitempty"`
	OutputIndex        int             `json:"output_index,omitempty"`
	Amount             decimal.Decimal `json:"amount,omitempty"`
	Senders            []string        `json:"senders,omitempty"`
	SendersThreshold   uint8           `json:"senders_threshold,omitempty"`
	Receivers          []string        `json:"receivers,omitempty"`
	ReceiversThreshold uint8           `json:"receivers_threshold,omitempty"`
	State              string          `json:"state,omitempty"`
	SignedBy           string          `json:"signed_by,omitempty"`
	SignedTx           string          `json:"signed_tx,omitempty"`
}

type CollectibleRequest struct {
	Type               string          `json:"type,omitempty"`
	CreatedAt          time.Time       `json:"created_at,omitempty"`
	UpdatedAt          time.Time       `json:"updated_at,omitempty"`
	RequestID          string          `json:"request_id,omitempty"`
	UserID             string          `json:"user_id,omitempty"`
	TokenID            string          `json:"token_id,omitempty"`
	Amount             decimal.Decimal `json:"amount,omitempty"`
	Senders            []string        `json:"senders,omitempty"`
	SendersThreshold   uint8           `json:"senders_threshold,omitempty"`
	Receivers          []string        `json:"receivers,omitempty"`
	ReceiversThreshold uint8           `json:"receivers_threshold,omitempty"`
	Signers            []string        `json:"signers,omitempty"`
	Action             string          `json:"action,omitempty"`
	State              string          `json:"state,omitempty"`
	TransactionHash    Hash            `json:"transaction_hash,omitempty"`
	RawTransaction     string          `json:"raw_transaction,omitempty"`
	CodeID             string          `json:"code_id"`
}

// ReadCollectibleOutputs return a list of collectibles outputs
func (c *Client) ReadCollectibleOutputs(ctx context.Context, members []string, threshold uint8, state string, offset time.Time, limit int) ([]*CollectibleOutput, error) {
	params := make(map[string]string)
	if !offset.IsZero() {
		params["offset"] = offset.UTC().Format(time.RFC3339Nano)
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	if state != "" {
		params["state"] = state
	}

	if len(members) > 0 {
		if threshold < 1 || int(threshold) > len(members) {
			return nil, errors.New("invalid members")
		}

		params["members"] = HashMembers(members)
		params["threshold"] = strconv.Itoa(int(threshold))
	}

	var outputs []*CollectibleOutput
	if err := c.Get(ctx, "/collectibles/outputs", params, &outputs); err != nil {
		return nil, err
	}

	return outputs, nil
}

// ReadCollectibleOutputs request with accessToken and returns a list of collectibles outputs
func ReadCollectibleOutputs(ctx context.Context, accessToken string, members []string, threshold uint8, state string, offset time.Time, limit int) ([]*CollectibleOutput, error) {
	return NewFromAccessToken(accessToken).ReadCollectibleOutputs(ctx, members, threshold, state, offset, limit)
}

func (c *Client) MakeCollectibleTransaction(
	ctx context.Context,
	output *CollectibleOutput,
	token *CollectibleToken,
	receivers []string,
	threshold uint8,
) (*Transaction, error) {
	tx := &Transaction{
		Version: TxVersion,
		Asset:   token.MixinID,
		Extra:   token.NFO,
		Inputs: []*Input{{
			Hash:  &output.TransactionHash,
			Index: output.OutputIndex,
		}},
	}

	ghostInputs, err := c.BatchReadGhostKeys(ctx, []*GhostInput{{
		Receivers: receivers,
		Index:     0,
		Hint:      output.OutputID,
	}})
	if err != nil {
		return nil, err
	}

	tx.Outputs = []*Output{ghostInputs[0].DumpOutput(threshold, output.Amount)}
	return tx, nil
}

// MakeCollectibleTransaction make collectible transaction with accessToken
func MakeCollectibleTransaction(
	ctx context.Context,
	accessToken string,
	output *CollectibleOutput,
	token *CollectibleToken,
	receivers []string,
	threshold uint8,
) (*Transaction, error) {
	return NewFromAccessToken(accessToken).MakeCollectibleTransaction(ctx, output, token, receivers, threshold)
}

// CreateCollectibleRequest create a collectibles request
func (c *Client) CreateCollectibleRequest(ctx context.Context, action, raw string) (*CollectibleRequest, error) {
	params := map[string]string{
		"action": action,
		"raw":    raw,
	}

	var req CollectibleRequest
	if err := c.Post(ctx, "/collectibles/requests", params, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// CreateCollectibleRequest create a collectibles request with accessToken
func CreateCollectibleRequest(ctx context.Context, accessToken, action, raw string) (*CollectibleRequest, error) {
	return NewFromAccessToken(accessToken).CreateCollectibleRequest(ctx, action, raw)
}

// SignCollectibleRequest sign a collectibles request
func (c *Client) SignCollectibleRequest(ctx context.Context, reqID, pin string) (*CollectibleRequest, error) {
	uri := "/collectibles/requests/" + reqID + "/sign"
	params := map[string]string{
		"pin": c.EncryptPin(pin),
	}

	var req CollectibleRequest
	if err := c.Post(ctx, uri, params, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// CancelCollectible cancel a collectibles request
func (c *Client) CancelCollectibleRequest(ctx context.Context, reqID string) error {
	uri := "/collectibles/requests/" + reqID + "/cancel"
	if err := c.Post(ctx, uri, nil, nil); err != nil {
		return err
	}

	return nil
}

// CancelCollectible cancel a collectibles request with accessToken
func CancelCollectibleRequest(ctx context.Context, accessToken, reqID string) error {
	return NewFromAccessToken(accessToken).CancelCollectibleRequest(ctx, reqID)
}

// UnlockCollectibleRequest unlock a collectibles request
func (c *Client) UnlockCollectibleRequest(ctx context.Context, reqID, pin string) error {
	var (
		uri    = "/collectibles/requests/" + reqID + "/unlock"
		params = map[string]string{
			"pin": c.EncryptPin(pin),
		}
	)
	if err := c.Post(ctx, uri, params, nil); err != nil {
		return err
	}

	return nil
}
