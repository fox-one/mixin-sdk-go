package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircle(t *testing.T) {
	store := newKeystoreFromEnv(t)
	c, err := NewFromKeystore(&store.Keystore)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	ctx := context.Background()
	name := RandomPin()

	circle, err := c.CreateCircle(ctx, CreateCircleParams{
		Name: name,
	})

	require.NoError(t, err)
	assert.Equal(t, circle.Name, name)

	t.Run("read circle", func(t *testing.T) {
		c, err := c.ReadCircle(ctx, circle.ID)
		require.NoError(t, err)

		assert.Equal(t, c.Name, circle.Name)
	})

	t.Run("update circle", func(t *testing.T) {
		newName := RandomPin()
		c, err := c.UpdateCircle(ctx, UpdateCircleParams{
			CircleID: circle.ID,
			Name:     newName,
		})

		require.NoError(t, err)
		assert.Equal(t, c.Name, newName)
	})

	t.Run("add user", func(t *testing.T) {
		app, err := c.ReadApp(ctx, c.ClientID)
		require.NoError(t, err)

		item, err := c.ManageCircle(ctx, ManageCircleParams{
			CircleID: circle.ID,
			Action:   CircleActionAdd,
			ItemType: CircleItemTypeUsers,
			ItemID:   app.CreatorID,
		})
		require.NoError(t, err)
		assert.Equal(t, item.UserID, app.CreatorID)
	})

	t.Run("list items", func(t *testing.T) {
		items, err := c.ListCircleItems(ctx, ListCircleItemsParams{
			CircleID: circle.ID,
			Limit:    10,
		})
		require.NoError(t, err)
		require.Len(t, items, 1)
	})

	t.Run("delete circle", func(t *testing.T) {
		err := c.DeleteCircle(ctx, circle.ID)
		require.NoError(t, err)
	})
}
