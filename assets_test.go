package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAsset(t *testing.T) {
	ctx := context.Background()
	store := newKeystoreFromEnv(t)

	c, err := NewFromKeystore(&store.Keystore)
	require.Nil(t, err, "init client")

	asset, err := c.ReadAsset(ctx, "c6d0c728-2624-429b-8e0d-d9d19b6592fa")
	require.Nil(t, err, "read asset")
	require.True(t, asset.DepositEntries != nil && len(asset.DepositEntries) > 0, "bitcoin missing segwit address")
}
