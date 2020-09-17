package mixin

import (
	"crypto/ed25519"
	"encoding/base64"
)

func GenerateEd25519Key() ed25519.PrivateKey {
	_, private, _ := ed25519.GenerateKey(nil)
	return private
}

var ed25519Encoding = &ed25519Encoder{}

type ed25519Encoder struct{}

func (enc *ed25519Encoder) EncodeToString(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func (enc *ed25519Encoder) DecodeString(s string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		b, err = base64.StdEncoding.DecodeString(s)
	}

	return b, err
}
