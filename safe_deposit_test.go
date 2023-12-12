package mixin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeDeposits(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	entries, err := dapp.SafeCreateDepositEntries(ctx, []string{dapp.ClientID}, 0, "b91e18ff-a9ae-3dc7-8679-e935d9a4b34b")
	require.NoError(err, "SafeCreateDepositEntries")
	require.NotEmpty(entries)
	{
		bts, _ := json.MarshalIndent(entries, "", "    ")
		t.Log(string(bts))
	}

	_, err = dapp.SafeListDeposits(ctx, entries[0], "", time.Time{}, 10)
	require.NoError(err, "SafeListDeposits")
}
