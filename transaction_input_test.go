package mixin

import (
	"context"
	"testing"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestBuildTransaction(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	t.Run("legacy-network", func(t *testing.T) {
	})

	t.Run("safe-network", func(t *testing.T) {
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

		addr := RequireNewMixAddress([]string{dapp.ClientID}, 1)
		outputs := []*TransactionOutputInput{
			{
				Address: *addr,
				Amount:  decimal.New(1, -8),
			},
		}
		if utxos[0].Amount.GreaterThan(decimal.New(1, -8)) {
			outputs = append(outputs, &TransactionOutputInput{
				Address: *addr,
				Amount:  utxos[0].Amount.Sub(decimal.New(1, -8)),
			})
		}

		tx, err := dapp.MakeSafeTransaction(
			ctx,
			uuid.Must(uuid.NewV4()).String(),
			utxos,
			outputs,
			nil,
			"TestSafeMakeTransaction",
		)
		require.NoError(err, "MakeSafeTransaction")

		raw, err := tx.Dump()
		require.NoError(err, "Dump")
		t.Log(raw)

		requests, err := dapp.SafeCreateTransactionRequest(ctx, []*SafeTransactionRequestInput{
			{
				RequestID:      uuidHash([]byte(utxos[0].OutputID + ":SafeCreateTransactionRequest")),
				RawTransaction: raw,
			},
		})
		require.NoError(err, "SafeCreateTransactionRequest")

		inputUtxos := make(map[mixinnet.Hash]map[uint64]*SafeUtxo, len(utxos))
		for _, utxo := range utxos {
			if m, ok := inputUtxos[utxo.TransactionHash]; ok {
				m[utxo.OutputIndex] = utxo
			} else {
				inputUtxos[utxo.TransactionHash] = map[uint64]*SafeUtxo{
					utxo.OutputIndex: utxo,
				}
			}
		}

		signedTx, err := SafeSignTransaction(
			ctx,
			*store.SpendKey,
			requests[0],
			inputUtxos,
		)
		require.NoError(err, "SafeSignTransaction")

		signedRaw, err := signedTx.Dump()
		require.NoError(err, "tx.Dump")

		requests1, err := dapp.SafeSubmitTransactionRequest(ctx, []*SafeTransactionRequestInput{
			{
				RequestID:      uuidHash([]byte(utxos[0].OutputID + ":SafeSubmitTransactionRequest")),
				RawTransaction: signedRaw,
			},
		})
		require.NoError(err, "SafeSubmitTransactionRequest")

		_, err = dapp.SafeReadTransactionRequest(ctx, requests1[0].RequestID)
		require.NoError(err, "SafeReadTransactionRequest")

		t.Log(requests1[0].TransactionHash)
	})

}
