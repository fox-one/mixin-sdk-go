package mixin

import (
	"context"
	"crypto/ed25519"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	Signer
	Verifier
	MessageLocker

	ClientID string
}

func newClient(id string) *Client {
	return &Client{
		ClientID:      id,
		Verifier:      NopVerifier(),
		MessageLocker: &messageLockNotSupported{},
	}
}

func NewFromKeystore(keystore *Keystore) (*Client, error) {
	auth, err := AuthFromKeystore(keystore)
	if err != nil {
		return nil, err
	}

	c := newClient(keystore.ClientID)
	c.Signer = auth

	if key, ok := auth.signKey.(ed25519.PrivateKey); ok {
		c.MessageLocker = &ed25519MessageLocker{
			sessionID: keystore.SessionID,
			key:       key,
		}
	}

	return c, nil
}

func NewFromAccessToken(accessToken string) *Client {
	c := newClient("")
	c.Signer = accessTokenAuth(accessToken)

	return c
}

func NewFromOauthKeystore(keystore *OauthKeystore) (*Client, error) {
	c := newClient(keystore.ClientID)

	auth, err := AuthFromOauthKeystore(keystore)
	if err != nil {
		return nil, err
	}

	c.Signer = auth
	c.Verifier = auth

	return c, nil
}

func (c *Client) Request(ctx context.Context) *resty.Request {
	ctx = WithVerifier(ctx, c.Verifier)
	ctx = WithSigner(ctx, c.Signer)
	return Request(ctx)
}

func (c *Client) Get(ctx context.Context, uri string, params map[string]string, resp interface{}) error {
	r, err := c.Request(ctx).SetQueryParams(params).Get(uri)
	if err != nil {
		if requestID := extractRequestID(r); requestID != "" {
			return WrapErrWithRequestID(err, requestID)
		}

		return err
	}

	return UnmarshalResponse(r, resp)
}

func (c *Client) Post(ctx context.Context, uri string, body interface{}, resp interface{}) error {
	r, err := c.Request(ctx).SetBody(body).Post(uri)
	if err != nil {
		if requestID := extractRequestID(r); requestID != "" {
			return WrapErrWithRequestID(err, requestID)
		}

		return err
	}

	return UnmarshalResponse(r, resp)
}
