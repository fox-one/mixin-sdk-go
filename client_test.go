package mixin

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixDecodeEd25519Key(t *testing.T) {
	const emptyPrivateKey = ""
	b, err := ed25519Encoding.DecodeString(emptyPrivateKey)
	assert.Nil(t, err, "decode empty string success")
	assert.False(t, len(b) == ed25519.PrivateKeySize)
}
