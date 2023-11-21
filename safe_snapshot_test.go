package mixin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeSnapshots(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	_, err = dapp.ReadSafeSnapshots(ctx, "", time.Time{}, "ASC", 100)
	require.NoError(err, "ReadSafeSnapshots")
}
