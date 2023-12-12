package mixin

import (
	"context"
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"golang.org/x/crypto/sha3"
)

func (c *Client) SafeMigrate(ctx context.Context, priv string, pin string) (*User, error) {
	privKey, err := mixinnet.KeyFromString(priv)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.Public()
	pinKey, err := mixinnet.KeyFromString(pin)
	if err != nil {
		return nil, err
	}

	sig := privKey.SignHash(sha3.Sum256([]byte(c.ClientID)))
	paras := map[string]interface{}{
		"public_key": pubKey.String(),
		"signature":  base64.RawURLEncoding.EncodeToString(sig[:]),
		"pin_base64": c.EncryptTipPin(pinKey, TIPSequencerRegister, c.ClientID, pubKey.String()),
	}

	var user User
	if err := c.Post(ctx, "/safe/users", paras, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
