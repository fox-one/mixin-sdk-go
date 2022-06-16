package mixin

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeKeystoreAndPinFromEnv(t *testing.T) (*Keystore, string) {
	env := "TEST_KEYSTORE_PATH"
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

func newKeystoreFromEnv(t *testing.T) *Keystore {
	keystore, _ := decodeKeystoreAndPinFromEnv(t)
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

func TestKeystoreAuth_SignTokenAt(t *testing.T) {
	auth := &KeystoreAuth{
		Keystore: &Keystore{
			ClientID:  newUUID(),
			SessionID: newUUID(),
		},
	}

	sig := SignRaw("GET", "/me", nil)
	requestID := newUUID()
	at := time.Now()
	exp := time.Minute

	t.Run("rsa", func(t *testing.T) {
		auth.signMethod = jwt.SigningMethodRS512
		auth.signKey, _ = rsa.GenerateKey(rand.Reader, 2048)

		assert.Equal(
			t,
			auth.SignTokenAt(sig, requestID, at, exp),
			auth.SignTokenAt(sig, requestID, at, exp),
			"token should be the same",
		)

		assert.Equal(
			t,
			auth.SignTokenAt(sig, requestID, at.Add(time.Hour), exp),
			auth.SignTokenAt(sig, requestID, at.Add(time.Hour), exp),
			"token should be the same",
		)
	})

	t.Run("ed25519", func(t *testing.T) {
		auth.signMethod = Ed25519SigningMethod
		auth.signKey = GenerateEd25519Key()

		assert.Equal(
			t,
			auth.SignTokenAt(sig, requestID, at, exp),
			auth.SignTokenAt(sig, requestID, at, exp),
			"token should be the same",
		)

		assert.Equal(
			t,
			auth.SignTokenAt(sig, requestID, at.Add(time.Hour), exp),
			auth.SignTokenAt(sig, requestID, at.Add(time.Hour), exp),
			"token should be the same",
		)
	})
}
