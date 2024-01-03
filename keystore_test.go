package mixin

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	SpenderKeystore struct {
		Keystore
		SpendKey mixinnet.Key `json:"spend_key"`
		Pin      string       `json:"pin"`
	}
)

func decodeKeystoreAndPinFromEnv(t *testing.T) *SpenderKeystore {
	ctx := context.Background()

	env := "TEST_KEYSTORE_PATH"
	path := os.Getenv(env)
	if path == "" {
		t.Logf("skip test, env %s not set", env)
		t.SkipNow()
	}

	f, err := os.Open(path)
	require.Nil(t, err, "open path: %v", path)

	defer f.Close()

	var store SpenderKeystore
	require.Nil(t, json.NewDecoder(f).Decode(&store), "decode keystore")

	client, err := NewFromKeystore(&store.Keystore)
	require.NoError(t, err, "init client")

	user, err := client.UserMe(ctx)
	require.NoError(t, err, "UserMe")

	if store.SpendKey.HasValue() {
		store.SpendKey, _ = mixinnet.ParseKeyWithPub(store.SpendKey.String(), user.SpendPublicKey)
	}

	if len(store.Pin) > 6 {
		pub, err := ed25519Encoding.DecodeString(user.TipKeyBase64)
		require.NoError(t, err, "decode tip key")

		pin, _ := mixinnet.ParseKeyWithPub(store.Pin, hex.EncodeToString(pub))
		store.Pin = pin.String()
	}

	return &store
}

func newKeystoreFromEnv(t *testing.T) *SpenderKeystore {
	return decodeKeystoreAndPinFromEnv(t)
}

func TestKeystoreAuth(t *testing.T) {
	s := newKeystoreFromEnv(t)

	auth, err := AuthFromKeystore(&s.Keystore)
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
		auth.signMethod = jwt.SigningMethodEdDSA
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
