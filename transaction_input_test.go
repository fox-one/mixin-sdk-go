package mixin

import (
	"context"
	"testing"

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

		b := NewSafeTransactionBuilder(utxos)
		b.Memo = "TestSafeMakeTransaction"

		tx, err := dapp.MakeTransaction(ctx, b, []*TransactionOutput{
			{
				Address: RequireNewMixAddress([]string{dapp.ClientID}, 1),
				Amount:  decimal.New(1, -8),
			},
		})
		require.NoError(err, "MakeTransaction")

		raw, err := tx.Dump()
		require.NoError(err, "Dump")
		t.Log(raw)

		request, err := dapp.SafeCreateTransactionRequest(ctx, &SafeTransactionRequestInput{
			RequestID:      uuidHash([]byte(utxos[0].OutputID + ":SafeCreateTransactionRequest")),
			RawTransaction: raw,
		})
		require.NoError(err, "SafeCreateTransactionRequest")
		err = SafeSignTransaction(
			tx,
			store.SpendKey,
			request.Views,
			0,
		)
		require.NoError(err, "SafeSignTransaction")

		signedRaw, err := tx.Dump()
		require.NoError(err, "tx.Dump")

		request1, err := dapp.SafeSubmitTransactionRequest(ctx, &SafeTransactionRequestInput{
			RequestID:      request.RequestID,
			RawTransaction: signedRaw,
		})
		require.NoError(err, "SafeSubmitTransactionRequest")

		_, err = dapp.SafeReadTransactionRequest(ctx, request1.RequestID)
		require.NoError(err, "SafeReadTransactionRequest")

		t.Log(request1.TransactionHash)
	})

}
