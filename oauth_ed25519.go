package mixin

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-resty/resty/v2"
)

type OauthKeystore struct {
	ClientID   string `json:"client_id,omitempty"`
	AuthID     string `json:"authorization_id,omitempty"`
	Scope      string `json:"scope,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	VerifyKey  string `json:"ed25519,omitempty"`
}

func AuthorizeEd25519(ctx context.Context, clientID, clientSecret string, code string, verifier string, privateKey ed25519.PrivateKey) (*OauthKeystore, error) {
	params := map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"code_verifier": verifier,
		"ed25519":       privateKey.Public(),
	}

	resp, err := Request(ctx).SetBody(params).Post("/oauth/token")
	if err != nil {
		return nil, err
	}

	var key OauthKeystore
	if err := UnmarshalResponse(resp, &key); err != nil {
		return nil, err
	}

	key.ClientID = clientID
	key.PrivateKey = base64.StdEncoding.EncodeToString(privateKey)

	return &key, nil
}

type OauthKeystoreAuth struct {
	*OauthKeystore
	signMethod jwt.SigningMethod
	signKey    interface{}
	verifyKey  interface{}
}

func AuthFromOauthKeystore(store *OauthKeystore) (*OauthKeystoreAuth, error) {
	auth := &OauthKeystoreAuth{
		OauthKeystore: store,
		signMethod:    Ed25519SigningMethod,
	}

	sign, err := base64.StdEncoding.DecodeString(store.PrivateKey)
	if err != nil {
		return nil, err
	}

	auth.signKey = (ed25519.PrivateKey)(sign)

	verify, err := base64.StdEncoding.DecodeString(store.VerifyKey)
	if err != nil {
		return nil, err
	}

	auth.verifyKey = (ed25519.PublicKey)(verify)

	return auth, nil
}

func (o *OauthKeystoreAuth) SignToken(signature, requestID string, exp time.Duration) string {
	jwtMap := jwt.MapClaims{
		"iss": o.ClientID,
		"aid": o.AuthID,
		"scp": o.Scope,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(exp).Unix(),
		"sig": signature,
		"jti": requestID,
	}

	token, err := jwt.NewWithClaims(Ed25519SigningMethod, jwtMap).SignedString(o.signKey)
	if err != nil {
		panic(err)
	}

	return token
}

func (o *OauthKeystoreAuth) EncryptPin(pin string) string {
	panic(errors.New("[oauth auth] encrypt pin: forbidden"))
}

func (o *OauthKeystoreAuth) Verify(resp *resty.Response) error {
	verifyToken := resp.Header().Get(xIntegrityToken)

	var claim struct {
		jwt.StandardClaims
		Sign string `json:"sig,omitempty"`
	}

	if _, err := jwt.ParseWithClaims(verifyToken, &claim, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*EdDSASigningMethod); !ok {
			return nil, jwt.ErrInvalidKeyType
		}

		return o.verifyKey, nil
	}); err != nil {
		return err
	}

	if expect, got := claim.Id, resp.Header().Get(xRequestID); expect != got {
		return fmt.Errorf("token.jti mismatch, expect %q but got %q", expect, got)
	}

	if expect, got := claim.Sign, SignResponse(resp); expect != got {
		return fmt.Errorf("token.sig mismatch, expect %q but got %q", expect, got)
	}

	return nil
}
