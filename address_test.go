package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	ctx := context.Background()
	store, pin := decodeKeystoreAndPinFromEnv(t)

	c, err := NewFromKeystore(store)
	require.Nil(t, err, "init client")

	input := CreateAddressInput{
		AssetID:     "c6d0c728-2624-429b-8e0d-d9d19b6592fa",
		Destination: "1M7aEv3BhcB2AtBTVZXVKwfW3p2We1bavT",
		Tag:         "",
		Label:       "my btc address",
	}

	address, err := c.CreateAddress(ctx, input, pin)
	require.Nil(t, err, "create address")

	t.Run("read address", func(t *testing.T) {
		_, err := c.ReadAddress(ctx, address.AddressID)
		assert.Nil(t, err, "read address")
	})

	t.Run("read addresses", func(t *testing.T) {
		addresses, err := c.ReadAddresses(ctx, input.AssetID)
		require.Nil(t, err, "read addresses")

		var ids []string
		for _, address := range addresses {
			ids = append(ids, address.AddressID)
		}

		assert.Contains(t, ids, address.AddressID, "should contain")
	})
}
