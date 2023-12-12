package mixin

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestSafeMultisigs(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

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
		{
			b := NewSafeTransactionBuilder(utxos)
			b.Memo = "Transfer To Multisig"
			b.Hint = newUUID()

			tx, err := dapp.MakeTransaction(ctx, b, []*TransactionOutput{
				{
					Address: RequireNewMixAddress([]string{dapp.ClientID, "6a00a4bc-229e-3c39-978a-91d2d6c382bf"}, 1),
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

			utxos = []*SafeUtxo{
				{
					OutputID:           newUUID(),
					KernelAssetID:      tx.Asset,
					TransactionHash:    *tx.Hash,
					OutputIndex:        0,
					Amount:             decimal.RequireFromString(tx.Outputs[0].Amount.String()),
					ReceiversThreshold: 1,
					Receivers:          []string{dapp.ClientID, "6a00a4bc-229e-3c39-978a-91d2d6c382bf"},
				},
			}
			time.Sleep(time.Second * 10)
		}
		var k uint16 = 0
		if strings.Compare(dapp.ClientID, "6a00a4bc-229e-3c39-978a-91d2d6c382bf") > 0 {
			k = 1
		}

		{
			b := NewSafeTransactionBuilder(utxos)
			b.Memo = "Transfer From Multisig"

			tx, err := dapp.MakeTransaction(ctx, b, []*TransactionOutput{
				{
					Address: RequireNewMixAddress([]string{dapp.ClientID}, 1),
					Amount:  decimal.New(1, -8),
				},
			})
			require.NoError(err, "MakeTransaction")

			raw, err := tx.Dump()
			require.NoError(err, "Dump")

			request, err := dapp.SafeCreateMultisigRequest(ctx, &SafeTransactionRequestInput{
				RequestID:      uuidHash([]byte(utxos[0].OutputID + ":SafeCreateMultisigRequest")),
				RawTransaction: raw,
			})
			require.NoError(err, "SafeCreateMultisigRequest")

			_, err = dapp.SafeUnlockMultisigRequest(ctx, request.RequestID)
			require.NoError(err, "SafeUnlockMultisigRequests")

			{
				hash, err := tx.TransactionHash()
				require.NoError(err, "TransactionHash")

				request1, err := dapp.SafeReadMultisigRequests(ctx, request.RequestID)
				require.NoError(err, "SafeReadMultisigRequests")
				require.Equal(hash.String(), request1.TransactionHash)
			}

			err = SafeSignTransaction(
				tx,
				store.SpendKey,
				request.Views,
				k,
			)
			require.NoError(err, "SafeSignTransaction")

			signedRaw, err := tx.Dump()
			require.NoError(err, "tx.Dump")
			t.Log("signed tx", signedRaw)

			request, err = dapp.SafeSignMultisigRequest(ctx, &SafeTransactionRequestInput{
				RequestID:      request.RequestID,
				RawTransaction: signedRaw,
			})
			require.NoError(err, "SafeSignMultisigRequests")

			_, err = dapp.SafeUnlockMultisigRequest(ctx, request.RequestID)
			require.Error(err, "SafeUnlockMultisigRequests Forbidden")
		}
	})
}
