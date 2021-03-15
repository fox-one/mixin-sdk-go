package mixin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	TxMethodSend = "sendrawtransaction"
	TxMethodGet  = "gettransaction"
)

func ReadConsensusInfo(ctx context.Context) (*ConsensusInfo, error) {
	var resp ConsensusInfo
	err := CallMixinNetRPC(ctx, &resp, "getinfo")
	return &resp, err
}

func SendRawTransaction(ctx context.Context, raw string) (*Transaction, error) {
	var tx Transaction
	if err := CallMixinNetRPC(ctx, &tx, TxMethodSend, raw); err != nil {
		if IsErrorCodes(err, InvalidOutputKey) {
			if tx, err := TransactionFromRaw(raw); err == nil {
				h, _ := tx.TransactionHash()
				if tx, err := GetTransaction(ctx, h); err == nil && tx.Asset.HasValue() {
					return tx, nil
				}
			}
		}
		return nil, err
	}

	return GetTransaction(ctx, *tx.Hash)
}

func GetTransaction(ctx context.Context, hash Hash) (*Transaction, error) {
	var tx Transaction
	if err := CallMixinNetRPC(ctx, &tx, TxMethodGet, hash); err != nil {
		return nil, err
	}
	return &tx, nil
}

func CallMixinNetRPC(ctx context.Context, resp interface{}, method string, params ...interface{}) error {
	r, err := MixinNetClientFromContext(ctx).R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"method": method,
			"params": params,
		}).Post("")
	if err != nil {
		return err
	}

	return UnmarshalMixinNetResponse(r, resp)
}

func DecodeMixinNetResponse(resp *resty.Response) ([]byte, error) {
	var body struct {
		Error string          `json:"error,omitempty"`
		Data  json.RawMessage `json:"data,omitempty"`
	}

	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		if resp.IsError() {
			return nil, createError(resp.StatusCode(), resp.StatusCode(), resp.Status())
		}

		return nil, createError(resp.StatusCode(), resp.StatusCode(), err.Error())
	}

	if body.Error != "" {
		return nil, parseMixinNetError(body.Error)
	}

	return body.Data, nil
}

func UnmarshalMixinNetResponse(resp *resty.Response, v interface{}) error {
	data, err := DecodeMixinNetResponse(resp)
	if err != nil {
		return err
	}

	if v != nil {
		return json.Unmarshal(data, v)
	}

	return nil
}

func parseMixinNetError(errMsg string) error {
	if strings.HasPrefix(errMsg, "invalid output key ") {
		return createError(202, InvalidOutputKey, errMsg)
	}

	if strings.HasPrefix(errMsg, "input locked for transaction ") {
		return createError(202, InputLocked, errMsg)
	}

	if strings.HasPrefix(errMsg, "invalid tx signature number ") ||
		strings.HasPrefix(errMsg, "invalid signature keys ") {
		return createError(202, InvalidSignature, errMsg)
	}

	return createError(202, 202, errMsg)
}
