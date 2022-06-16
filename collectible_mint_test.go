package mixin

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCollectibleTokenID(t *testing.T) {
	collection := "10b44d45-8871-4ce3-aa2e-ff09af519f71"
	token := int64(14)
	tokenID := GenerateCollectibleTokenID(collection, token)

	store := newKeystoreFromEnv(t)
	client, err := NewFromKeystore(store)
	require.NoError(t, err)

	ctx := context.Background()
	cToken, err := client.ReadCollectiblesToken(ctx, tokenID)
	require.NoError(t, err)
	assert.Equal(t, collection, cToken.CollectionID)
	assert.Equal(t, strconv.Itoa(int(token)), cToken.Token)
}
