package mixin

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newKeystoreFromTestData(t *testing.T) *Keystore {
	path := "./testdata/keystore.json"
	f, err := os.Open(path)
	require.Nil(t, err, "open path: %v", path)

	defer f.Close()

	var store Keystore
	require.Nil(t, json.NewDecoder(f).Decode(&store), "decode keystore")

	return &store
}

func TestKeystoreAuth(t *testing.T) {
	s := newKeystoreFromTestData(t)

	auth, err := AuthFromKeystore(s)
	require.Nil(t, err, "auth from keystore")

	sig := SignRaw("GET", "/me", nil)
	token := auth.SignToken(sig, newRequestID(), time.Minute)

	me, err := UserMe(context.TODO(), token)
	require.Nil(t, err, "UserMe")

	assert.Equal(t, s.ClientID, me.UserID, "client id should be same")
}

func TestEd25519KeystoreAuth(t *testing.T) {
	path := "./testdata/keystore_ed25519.json"
	f, err := os.Open(path)
	require.Nil(t, err, "open path: %v", path)

	defer f.Close()

	var store Keystore
	require.Nil(t, json.NewDecoder(f).Decode(&store), "decode keystore")

	auth, err := AuthEd25519FromKeystore(&store)
	require.Nil(t, err, "auth from keystore")

	sig := SignRaw("GET", "/me", nil)
	token := auth.SignToken(sig, newRequestID(), time.Minute)

	me, err := UserMe(context.TODO(), token)
	require.Nil(t, err, "UserMe")

	assert.Equal(t, store.ClientID, me.UserID, "client id should be same")
}

func decodePinFromTestData(t *testing.T) string {
	path := "./testdata/keystore.json"
	f, err := os.Open(path)
	require.Nil(t, err, "open path: %v", path)

	defer f.Close()

	var store struct {
		Pin string `json:"pin,omitempty"`
	}
	_ = json.NewDecoder(f).Decode(&store)

	return store.Pin
}
