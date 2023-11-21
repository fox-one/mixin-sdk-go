package mixin

import (
	"context"
	"fmt"

	"filippo.io/edwards25519"
)

func SafeSignTransaction(ctx context.Context, spendKey Key, request *SafeTransactionRequest, inputUtxos map[Hash]map[uint64]*SafeUtxo) (*Transaction, error) {
	y, err := spendKey.ToScalar()
	if err != nil {
		return nil, err
	}

	tx, err := TransactionFromRaw(request.RawTransaction)
	if err != nil {
		return nil, err
	}

	txHash, err := tx.TransactionHash()
	if err != nil {
		return nil, err
	}

	if tx.Signatures == nil {
		tx.Signatures = make([]map[uint16]*Signature, len(tx.Inputs))
		for i := range tx.Inputs {
			tx.Signatures[i] = map[uint16]*Signature{}
		}
	}
	viewOffset := 0
	for i, input := range tx.Inputs {
		utxos, ok := inputUtxos[*input.Hash]
		if !ok {
			inputTx, err := GetTransaction(ctx, *input.Hash)
			if err != nil {
				return nil, err
			}
			utxos = make(map[uint64]*SafeUtxo, len(inputTx.Outputs))
			for i, output := range inputTx.Outputs {
				utxos[uint64(i)] = &SafeUtxo{
					Mask: output.Mask,
					Keys: output.Keys,
				}
			}
			inputUtxos[*input.Hash] = utxos
		}

		utxo, ok := utxos[input.Index]
		if !ok {
			return nil, fmt.Errorf("utxo (%v : %d) not found", input.Hash, input.Index)
		}

		keysFilter := make(map[string]int)
		for i, k := range utxo.Keys {
			keysFilter[k.String()] = i
		}

		for offset := 0; offset < len(utxo.Keys); offset++ {
			view := request.Views[offset]
			x, err := view.ToScalar()
			if err != nil {
				panic(err)
			}
			t := edwards25519.NewScalar().Add(x, y)
			var key Key
			copy(key[:], t.Bytes())

			k, found := keysFilter[key.Public().String()]
			if !found {
				panic("invalid public key for the input")
			}

			sig := key.SignHash(txHash)
			tx.Signatures[i][uint16(k)] = &sig
		}

		viewOffset += len(utxo.Keys)
	}
	return tx, nil
}
