package mixin

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestSafeGhostKeys(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	ghostKeys, err := dapp.SafeCreateGhostKeys(ctx, []*GhostInput{
		{
			Receivers: []string{dapp.ClientID},
			Hint:      uuid.Must(uuid.NewV4()).String(),
		},
	})
	require.NoError(err, "SafeCreateGhostKeys")
	require.NotEmpty(ghostKeys)
	require.Equal(1, len(ghostKeys[0].Keys))
}
