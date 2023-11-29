package mixinnet

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyFromString(t *testing.T) {
	msg := []byte("sign message")

	for i := 0; i < 2000; i++ {
		priv := GenerateEd25519Key()
		pub := priv.Public().(ed25519.PublicKey)

		key, err := KeyFromString(hex.EncodeToString(priv))
		require.NoError(t, err, "KeyFromString(%s)", hex.EncodeToString(priv))
		pubKey := key.Public()
		require.True(t, pubKey.CheckKey(), "CheckKey")
		require.True(t, bytes.Equal(pub[:], pubKey[:]), "public Key of (%s) not matched: %s != %s", key, pubKey, hex.EncodeToString(pub))

		{
			sigBytes, err := priv.Sign(rand.Reader, msg, &ed25519.Options{})
			require.Nil(t, err)

			var sig Signature
			copy(sig[:], sigBytes)
			require.True(t, pubKey.Verify(msg, sig))
			require.True(t, ed25519.Verify(pub, msg, sigBytes))
		}

		{
			sig := key.Sign(msg)
			sigBytes := make([]byte, len(sig))
			copy(sigBytes, sig[:])
			require.True(t, pubKey.Verify(msg, sig))
			require.True(t, ed25519.Verify(pub, msg, sigBytes))
		}
	}
}
