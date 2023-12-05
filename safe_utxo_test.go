package mixin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeUtxo(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	utxos, err := dapp.SafeListUtxos(ctx, SafeListUtxoOption{
		Members: []string{dapp.ClientID},
		Limit:   50,
		Order:   "ASC",
		State:   SafeUtxoStateUnspent,
	})
	require.NoError(err, "SafeListUtxos")

	bts, _ := json.MarshalIndent(utxos, "", "    ")
	t.Log(string(bts))
}
