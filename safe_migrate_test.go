package mixin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/stretchr/testify/require"
)

func TestSafeMigrate(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	store := newKeystoreFromEnv(t)
	dapp, err := NewFromKeystore(&store.Keystore)
	require.NoError(err, "init bot client")

	priv := GenerateEd25519Key()
	_, keystore, err := dapp.CreateUser(ctx, priv, "name-ed25519")
	require.NoError(err, "create a user with a Ed25519 key")

	subClient, err := NewFromKeystore(keystore)
	require.NoError(err, "Ed25519 user client")

	pin := mixinnet.GenerateKey(rand.Reader)
	err = subClient.ModifyPin(context.TODO(), "", pin.Public().String())
	require.NoError(err, "the Ed25519 user modifies pin")
	require.NoError(subClient.VerifyPin(ctx, pin.String()), "the Ed25519 user verify pin")

	spendKey := mixinnet.GenerateKey(rand.Reader)
	user, err := subClient.SafeMigrate(ctx, spendKey.String(), pin.String())
	require.NoError(err, "migrate failed")
	require.Equal(subClient.ClientID, user.UserID)
	require.True(user.HasSafe)

	bts, _ := json.Marshal(keystore)
	t.Log("new keystore", string(bts))
	t.Log("pin", pin)
	t.Log("spend key", spendKey)
}
