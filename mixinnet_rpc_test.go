package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTransaction(t *testing.T) {
	ctx := context.Background()
	hash, err := HashFromString("b6c4574730650502fda85a2a6bcb442f027a72b0d1ed25021834ad39e7423586")
	require.Nil(t, err)

	tx, err := GetTransaction(ctx, hash)
	require.Nil(t, err)

	hash1, err := tx.TransactionHash()
	require.Nil(t, err)
	require.Equal(t, hash[:], hash1[:], "hash not matched: %v != %v", hash, hash1)
}
