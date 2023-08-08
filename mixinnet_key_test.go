package mixin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyFromString(t *testing.T) {
	for i := 0; i < 2000; i++ {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err, "ed25519.GenerateKey")

		key, err := KeyFromString(hex.EncodeToString(priv))
		require.NoError(t, err, "KeyFromString", hex.EncodeToString(priv))
		require.True(t, key.CheckScalar(), "CheckScalar")
		pubKey := key.Public()
		require.True(t, pubKey.CheckKey(), "CheckKey")
		require.True(t, bytes.Equal(pub, pubKey[:]), "Public Key not matched", key, pubKey, hex.EncodeToString(pub))
	}
}
