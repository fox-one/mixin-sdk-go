package mixin

import (
	"context"

	"github.com/go-resty/resty/v2"
)

func (c *Client) CallMixinNetRPC(ctx context.Context, method string, params ...interface{}) (*resty.Response, error) {
	return c.Request(ctx).
		SetBody(map[string]interface{}{
			"method": method,
			"params": params,
		}).Post("/external/proxy")
}

func (c *Client) ReadConsensusInfo(ctx context.Context) (*ConsensusInfo, error) {
	r, err := c.callMixinNetRPC(ctx, "getinfo")
	if err != nil {
		return nil, err
	}

	var info ConsensusInfo
	if err := UnmarshalMixinNetResponse(r, &info); err != nil {
		return nil, err
	}
	return &info, err
}

func (c *Client) SendRawTransaction(ctx context.Context, raw string) (*Hash, error) {
	var tx Transaction
	r, err := c.callMixinNetRPC(ctx, TxMethodSend, raw)
	if err != nil {
		return nil, err
	}

	if err := UnmarshalMixinNetResponse(r, &tx); err != nil {
		return nil, err
	}

	return tx.Hash, nil
}

func (c *Client) GetRawTransaction(ctx context.Context, hash Hash) (*Transaction, error) {
	var tx Transaction

	r, err := c.callMixinNetRPC(ctx, TxMethodGet, hash)
	if err != nil {
		return nil, err
	}

	if err := UnmarshalMixinNetResponse(r, &tx); err != nil {
		return nil, err
	}

	return &tx, nil
}
