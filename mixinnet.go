package mixin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

var mixinNetClient = resty.New().
	SetHeader("Content-Type", "application/json").
	SetHostURL(DefaultMixinNetHost).
	SetTimeout(10 * time.Second)

func SendRawTransaction(ctx context.Context, raw string) (*Transaction, error) {
	var tx Transaction
	if err := callMixinNetRPC(ctx, map[string]interface{}{
		"method": "sendrawtransaction",
		"params": []interface{}{raw},
	}, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	var tx Transaction
	if err := callMixinNetRPC(ctx, map[string]interface{}{
		"method": "gettransaction",
		"params": []interface{}{hash},
	}, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func callMixinNetRPC(ctx context.Context, params interface{}, resp interface{}) error {
	r, err := mixinNetClient.R().SetContext(ctx).SetBody(params).Post("")
	if err != nil {
		return err
	}

	return UnmarshalMixinNetResponse(r, resp)
}

func DecodeMixinNetResponse(resp *resty.Response) ([]byte, error) {
	var body struct {
		Error interface{}     `json:"error,omitempty"`
		Data  json.RawMessage `json:"data,omitempty"`
	}

	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		if resp.IsError() {
			return nil, createError(resp.StatusCode(), resp.StatusCode(), resp.Status())
		}

		return nil, createError(resp.StatusCode(), resp.StatusCode(), err.Error())
	}

	if body.Error != nil {
		return nil, fmt.Errorf("MIXIN NET RPC ERROR %s", body.Error)
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
