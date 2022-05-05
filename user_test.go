package mixin

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	require := require.New(t)

	store := newKeystoreFromEnv(t)

	botClient, err := NewFromKeystore(store)
	require.NoError(err, "init bot client")

	// create a user with a RSA key
	rsaPriKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	user, keystore, err := botClient.CreateUser(context.TODO(), rsaPriKey, "name-rsa")
	require.NoError(err, "create a user with a RSA key")

	rsaUserClient, err := NewFromKeystore(keystore)
	require.NoError(err, "RSA user client")
	me, err := rsaUserClient.UserMe(context.TODO())
	require.NoError(err, "read the RSA user")
	require.Equal(me.UserID, user.UserID, "user ids should be same")
	err = rsaUserClient.ModifyPin(context.TODO(), "", "111111")
	require.NoError(err, "the RSA user modifies pin")

	ed25519PriKey := GenerateEd25519Key()
	user, keystore, err = botClient.CreateUser(context.TODO(), ed25519PriKey, "name-ed25519")
	require.NoError(err, "create a user with a Ed25519 key")

	ed25519UserClient, err := NewFromKeystore(keystore)
	require.NoError(err, "Ed25519 user client")
	me, err = ed25519UserClient.UserMe(context.TODO())
	require.NoError(err, "read the Ed25519 user")
	require.Equal(me.UserID, user.UserID, "user ids should be same")
	err = ed25519UserClient.ModifyPin(context.TODO(), "", "222222")
	require.NoError(err, "the Ed25519 user modifies pin")
}
