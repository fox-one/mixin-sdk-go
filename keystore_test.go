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

func decodeKeystoreAndPinFromEnv(t *testing.T, envName ...string) (*Keystore, string) {
	env := "TEST_KEYSTORE_PATH"
	if len(envName) > 0 {
		env = envName[0]
	}

	path := os.Getenv(env)
	if path == "" {
		t.SkipNow()
	}

	f, err := os.Open(path)
	require.Nil(t, err, "open path: %v", path)

	defer f.Close()

	var store struct {
		Keystore
		Pin string `json:"pin,omitempty"`
	}
	require.Nil(t, json.NewDecoder(f).Decode(&store), "decode keystore")

	return &store.Keystore, store.Pin
}

func newKeystoreFromEnv(t *testing.T, envName ...string) *Keystore {
	keystore, _ := decodeKeystoreAndPinFromEnv(t, envName...)
	return keystore
}

func TestKeystoreAuth(t *testing.T) {
	s := newKeystoreFromEnv(t)

	auth, err := AuthFromKeystore(s)
	require.Nil(t, err, "auth from keystore")

	sig := SignRaw("GET", "/me", nil)
	token := auth.SignToken(sig, newRequestID(), time.Minute)

	me, err := UserMe(context.TODO(), token)
	require.Nil(t, err, "UserMe")

	assert.Equal(t, s.ClientID, me.UserID, "client id should be same")
}

func TestEd25519KeystoreAuth(t *testing.T) {
	store := newKeystoreFromEnv(t, "TEST_KEYSTORE_ED25519_PATH")

	auth, err := AuthEd25519FromKeystore(store)
	require.Nil(t, err, "auth from keystore")

	sig := SignRaw("GET", "/me", nil)
	token := auth.SignToken(sig, newRequestID(), time.Minute)

	me, err := UserMe(context.TODO(), token)
	require.Nil(t, err, "UserMe")

	assert.Equal(t, store.ClientID, me.UserID, "client id should be same")
}
