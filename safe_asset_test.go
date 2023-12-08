package mixin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeAssets(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	assets, err := dapp.SafeReadAssets(ctx)
	require.NoError(err, "ReadSafeAssets")
	require.NotEmpty(assets, "/safe/assets return empty")

	asset, err := dapp.SafeReadAsset(ctx, "965e5c6e-434c-3fa9-b780-c50f43cd955c")
	require.NoError(err, "ReadSafeAsset")
	require.NotNil(asset, "/safe/asset/:id return nil")

	bts, _ := json.MarshalIndent(asset, "", "  ")
	t.Log(string(bts))
}
