package mixin

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestSafeMakeTransaction(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	utxos, err := dapp.SafeListUtxos(ctx, SafeListUtxoOption{
		Members: []string{dapp.ClientID},
		Limit:   1,
		State:   SafeUtxoStateUnspent,
	})
	require.NoError(err, "SafeListUtxos")
	if len(utxos) == 0 {
		t.Log("empty unspent utxo")
		return
	}

	input := SafeTransactionInput{
		Memo:   "TestSafeMakeTransaction",
		Inputs: utxos,
		Outputs: []SafeTransactionOutput{
			{
				Receivers: []string{"6a00a4bc-229e-3c39-978a-91d2d6c382bf"},
				Threshold: 1,
				Amount:    decimal.New(1, -8),
			},
		},
		Hint: uuid.Must(uuid.NewV4()).String(),
	}

	tx, err := dapp.SafeBuildTransaction(ctx, &input)
	require.NoError(err, "SafeBuildTransaction")

	txBytes, err := tx.DumpTransactionPayload()
	require.NoError(err, "DumpTransactionPayload")

	requests, err := dapp.SafeCreateTransactionRequest(ctx, []*SafeTransactionRequestInput{
		{
			RequestID:      uuidHash([]byte(utxos[0].OutputID + ":SafeCreateTransactionRequest")),
			RawTransaction: hex.EncodeToString(txBytes),
		},
	})
	require.NoError(err, "SafeCreateTransactionRequest")

	inputUtxos := make(map[Hash]map[uint64]*SafeUtxo, len(utxos))
	for _, utxo := range utxos {
		if m, ok := inputUtxos[utxo.TransactionHash]; ok {
			m[utxo.OutputIndex] = utxo
		} else {
			inputUtxos[utxo.TransactionHash] = map[uint64]*SafeUtxo{
				utxo.OutputIndex: utxo,
			}
		}
	}

	tx, err = SafeSignTransaction(
		ctx,
		*store.SpendKey,
		requests[0],
		inputUtxos,
	)
	require.NoError(err, "SafeSignTransaction")

	raw, err := tx.DumpTransaction()
	require.NoError(err, "DumpTransaction")

	requests1, err := dapp.SafeSubmitTransactionRequest(ctx, []*SafeTransactionRequestInput{
		{
			RequestID:      uuidHash([]byte(utxos[0].OutputID + ":SafeSubmitTransactionRequest")),
			RawTransaction: raw,
		},
	})
	require.NoError(err, "SafeSubmitTransactionRequest")

	_, err = dapp.SafeReadTransactionRequest(ctx, requests1[0].RequestID)
	require.NoError(err, "SafeReadTransactionRequest")
}
