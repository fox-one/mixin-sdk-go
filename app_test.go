package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateApp(t *testing.T) {
	ctx := context.Background()
	store := newKeystoreFromEnv(t)

	c, err := NewFromKeystore(store)
	require.Nil(t, err, "init client")

	app, err := c.ReadApp(ctx, store.ClientID)
	require.Nil(t, err, "read app")
	t.Log("old name", app.Name)

	name := app.Name
	newName := "new name"

	req := UpdateAppRequest{Name: newName}
	newApp, err := c.UpdateApp(ctx, app.AppID, req)
	require.Nil(t, err, "update app")
	t.Log("new name", newApp.Name)
	require.Equal(t, newApp.Name, newName, "name should changed")

	// restore name
	req.Name = name
	_, err = c.UpdateApp(ctx, app.AppID, req)
	require.Nil(t, err, "update app")
}
