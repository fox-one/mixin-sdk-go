package mixin

import (
	"context"
	"fmt"

	"filippo.io/edwards25519"
	"github.com/fox-one/mixin-sdk-go/mixinnet"
)

func SafeSignTransaction(ctx context.Context, spendKey mixinnet.Key, request *SafeTransactionRequest, inputUtxos map[mixinnet.Hash]map[uint64]*SafeUtxo) (*mixinnet.Transaction, error) {
	y, err := spendKey.ToScalar()
	if err != nil {
		return nil, err
	}

	tx, err := mixinnet.TransactionFromRaw(request.RawTransaction)
	if err != nil {
		return nil, err
	}

	txHash, err := tx.TransactionHash()
	if err != nil {
		return nil, err
	}

	if tx.Signatures == nil {
		tx.Signatures = make([]map[uint16]*mixinnet.Signature, len(tx.Inputs))
		for i := range tx.Inputs {
			tx.Signatures[i] = map[uint16]*mixinnet.Signature{}
		}
	}
	viewOffset := 0
	client := mixinnet.DefaultClient(tx.Version >= mixinnet.TxVersionHashSignature)
	for i, input := range tx.Inputs {
		utxos, ok := inputUtxos[mixinnet.Hash(*input.Hash)]
		if !ok {
			inputTx, err := client.GetTransaction(ctx, *input.Hash)
			if err != nil {
				return nil, err
			}
			utxos = make(map[uint64]*SafeUtxo, len(inputTx.Outputs))
			for i, output := range inputTx.Outputs {
				keys := make([]mixinnet.Key, len(output.Keys))
				for i, k := range output.Keys {
					keys[i] = mixinnet.Key(k)
				}
				utxos[uint64(i)] = &SafeUtxo{
					Mask: mixinnet.Key(output.Mask),
					Keys: keys,
				}
			}
			inputUtxos[mixinnet.Hash(*input.Hash)] = utxos
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
			var key mixinnet.Key
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
