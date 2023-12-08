package mixin

import (
	"context"
	"encoding/json"
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

	snapshots, err := dapp.ReadSafeSnapshots(ctx, "", time.Time{}, "ASC", 10)
	require.NoError(err, "ReadSafeSnapshots")

	bts, _ := json.MarshalIndent(snapshots, "", "  ")
	t.Log(string(bts))
}
