package mixinnet

import (
	"context"
	"errors"
)

const (
	TxMethodSend    = "sendrawtransaction"
	TxMethodGet     = "gettransaction"
	TxMethodGetUtxo = "getutxo"
)

func (c *Client) ReadConsensusInfo(ctx context.Context) (*ConsensusInfo, error) {
	var resp ConsensusInfo
	err := c.CallMixinNetRPC(ctx, &resp, "getinfo")
	return &resp, err
}

func (c *Client) SendRawTransaction(ctx context.Context, raw string) (*Transaction, error) {
	var tx Transaction
	if err := c.CallMixinNetRPC(ctx, &tx, TxMethodSend, raw); err != nil {
		if IsErrorCodes(err, InvalidOutputKey) {
			if tx, err := TransactionFromRaw(raw); err == nil {
				h, _ := tx.TransactionHash()
				if tx, err := c.GetTransaction(ctx, h); err == nil && tx.Asset.HasValue() {
					return tx, nil
				}
			}
		}
		return nil, err
	} else if tx.Hash == nil {
		return nil, errors.New("nil transaction hash")
	}

	return c.GetTransaction(ctx, *tx.Hash)
}

func (c *Client) GetTransaction(ctx context.Context, hash Hash) (*Transaction, error) {
	var tx Transaction
	if err := c.CallMixinNetRPC(ctx, &tx, TxMethodGet, hash); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (c *Client) GetUTXO(ctx context.Context, hash Hash, outputIndex uint8) (*UTXO, error) {
	var utxo UTXO
	if err := c.CallMixinNetRPC(ctx, &utxo, TxMethodGetUtxo, hash, outputIndex); err != nil {
		return nil, err
	}
	return &utxo, nil
}
