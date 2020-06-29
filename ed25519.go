package mixin

import (
	"crypto/ed25519"
	"encoding/base64"
)

func GenerateEd25519Key() ed25519.PrivateKey {
	_, private, _ := ed25519.GenerateKey(nil)
	return private
}

var ed25519Encoding = base64.RawURLEncoding
