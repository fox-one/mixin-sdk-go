package mixin

import (
	"context"
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

	assets, err := dapp.SafeReadAssets(ctx)
	require.NoError(err, "ReadSafeAssets")
	require.NotEmpty(assets, "/safe/assets return empty")

	entries, err := dapp.SafeCreateDepositEntries(ctx, []string{dapp.ClientID}, 0, assets[0].AssetID)
	require.NoError(err, "SafeCreateDepositEntries")
	require.NotEmpty(entries)

	_, err = dapp.SafeListDeposits(ctx, entries[0], "", time.Time{}, 10)
	require.NoError(err, "SafeListDeposits")
}
