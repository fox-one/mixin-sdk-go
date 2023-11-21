package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("legacy-network", func(t *testing.T) {
		UseLegacyMixinNetHosts()
		hash, err := HashFromString("f91402e0b55dc1555e7cea8bc497d6d61b6ec838c14a3f2d406893a4507ccf9e")
		require.Nil(t, err)

		tx, err := GetTransaction(ctx, hash)
		require.Nil(t, err)

		tx.Hash = nil
		hash1, err := tx.TransactionHash()
		require.Nil(t, err)
		require.Equal(t, hash[:], hash1[:], "hash not matched: %v != %v", hash, hash1)
	})

	t.Run("safe-network", func(t *testing.T) {
		UseSafeMixinNetHosts()
		hash, err := HashFromString("b6c4574730650502fda85a2a6bcb442f027a72b0d1ed25021834ad39e7423586")
		require.Nil(t, err)

		tx, err := GetTransaction(ctx, hash)
		require.Nil(t, err)

		hash1, err := tx.TransactionHash()
		require.Nil(t, err)
		require.Equal(t, hash[:], hash1[:], "hash not matched: %v != %v", hash, hash1)
	})
}
