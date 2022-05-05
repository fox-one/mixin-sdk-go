package mixin

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestClient_VerifyPayment(t *testing.T) {
	ctx := context.Background()
	store := newKeystoreFromEnv(t)

	c, err := NewFromKeystore(store)
	require.Nil(t, err, "init client")

	t.Run("test normal payment", func(t *testing.T) {
		payment, err := c.VerifyPayment(ctx, TransferInput{
			AssetID:    "965e5c6e-434c-3fa9-b780-c50f43cd955c",
			OpponentID: "d33ec557-b14c-403f-9f7a-08ed0f5866d4",
			Amount:     decimal.NewFromInt(1),
			TraceID:    newUUID(),
			Memo:       "memo",
		})

		require.NoError(t, err)
		require.NotNil(t, payment.Recipient)
		require.NotNil(t, payment.Asset)
		require.Equal(t, PaymentStatusPending, payment.Status)
	})

	t.Run("test multisig payment", func(t *testing.T) {
		input := TransferInput{
			AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c",
			Amount:  decimal.NewFromInt(1),
			TraceID: newUUID(),
			Memo:    "memo",
		}

		input.OpponentMultisig.Threshold = 1
		input.OpponentMultisig.Receivers = []string{
			store.ClientID,
			"d33ec557-b14c-403f-9f7a-08ed0f5866d4",
		}
		payment, err := c.VerifyPayment(ctx, input)

		require.NoError(t, err)
		require.Nil(t, payment.Recipient)
		require.Nil(t, payment.Asset)
		require.NotEmpty(t, payment.CodeID)
		require.Equal(t, PaymentStatusPending, payment.Status)

		t.Log(URL.Codes(payment.CodeID))
	})
}
