package mixin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fox-one/mixin-sdk-go/mixinnet"
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
		input := mixinnet.TransactionInput{
			TxVersion: mixinnet.TxVersionHashSignature,
			Memo:      "TestSafeMakeTransaction",
			Inputs: []*mixinnet.InputUTXO{
				{
					Input: mixinnet.Input{
						Hash:  &utxos[0].TransactionHash,
						Index: utxos[0].OutputIndex,
					},
					Asset:  utxos[0].Asset,
					Amount: utxos[0].Amount,
				},
			},
			Hint: uuid.Must(uuid.NewV4()).String(),
		}

		addr, err := NewMixAddress([]string{"6a00a4bc-229e-3c39-978a-91d2d6c382bf"}, byte(1))
		require.NoError(err, "NewMixAddress")

		{
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

			require.NoError(dapp.AppendOutputsToInput(ctx, &input, outputs), "AppendOutputsToInput")
		}

		tx, err := input.Build()
		require.NoError(err, "TransactionInput.Build")

		raw, err := tx.Dump()
		require.NoError(err, "Dump")

		{
			bts, _ := json.MarshalIndent(tx, "", "  ")
			t.Log(string(bts))
		}
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
	})

}
