package mixin

import (
	"context"
	"errors"
	"time"
)

func AuthorizeToken(ctx context.Context, clientID, clientSecret string, code string, verifier string) (string, string, error) {
	params := map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"code_verifier": verifier,
	}

	resp, err := Request(ctx).SetBody(params).Post("/oauth/token")
	if err != nil {
		return "", "", err
	}

	var body struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
	}

	err = UnmarshalResponse(resp, &body)
	return body.AccessToken, body.Scope, err
}

type accessTokenAuth string

func (a accessTokenAuth) SignToken(signature, requestID string, exp time.Duration) string {
	return string(a)
}

func (a accessTokenAuth) EncryptPin(pin string) string {
	panic(errors.New("[access token auth] encrypt pin: forbidden"))
}
