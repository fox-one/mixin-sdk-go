package mixin

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"
)

type User struct {
	UserID         string    `json:"user_id,omitempty"`
	IdentityNumber string    `json:"identity_number,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	FullName       string    `json:"full_name,omitempty"`
	Biography      string    `json:"biography,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	Relationship   string    `json:"relationship,omitempty"`
	MuteUntil      time.Time `json:"mute_until,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	IsVerified     bool      `json:"is_verified,omitempty"`
	SessionID      string    `json:"session_id,omitempty"`
	PinToken       string    `json:"pin_token,omitempty"`
	CodeID         string    `json:"code_id,omitempty"`
	CodeURL        string    `json:"code_url,omitempty"`
	HasPin         bool      `json:"has_pin,omitempty"`

	App *App `json:"app,omitempty"`
}

func (c *Client) UserMe(ctx context.Context) (*User, error) {
	var user User
	if err := c.Get(ctx, "/me", nil, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func UserMe(ctx context.Context, accessToken string) (*User, error) {
	return NewFromAccessToken(accessToken).UserMe(ctx)
}

func (c *Client) ReadUser(ctx context.Context, userIdOrIdentityNumber string) (*User, error) {
	uri := fmt.Sprintf("/users/%s", userIdOrIdentityNumber)

	var user User
	if err := c.Get(ctx, uri, nil, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) ReadUsers(ctx context.Context, ids ...string) ([]*User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var users []*User
	if err := c.Get(ctx, "/users/fetch", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) ReadFriends(ctx context.Context) ([]*User, error) {
	var users []*User
	if err := c.Get(ctx, "/friends", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// deprecated. Use ReadUser() instead
func (c *Client) SearchUser(ctx context.Context, identityNumber string) (*User, error) {
	return c.ReadUser(ctx, identityNumber)
}

func (c *Client) CreateUser(ctx context.Context, key *rsa.PrivateKey, fullname string) (*User, *Keystore, error) {
	pub, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return nil, nil, err
	}

	paras := map[string]interface{}{
		"session_secret": base64.StdEncoding.EncodeToString(pub),
		"full_name":      fullname,
	}

	var user User
	if err := c.Post(ctx, "/users", paras, &user); err != nil {
		return nil, nil, err
	}

	return &user, newKeystoreFromUser(&user, key), nil
}

func newKeystoreFromUser(user *User, privateKey *rsa.PrivateKey) *Keystore {
	pk := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return &Keystore{
		ClientID:   user.UserID,
		SessionID:  user.SessionID,
		PinToken:   user.PinToken,
		PrivateKey: string(pk),
	}
}

func (c *Client) ModifyProfile(ctx context.Context, fullname, avatarBase64 string) (*User, error) {
	params := map[string]interface{}{}
	if fullname != "" {
		params["full_name"] = fullname
	}
	if avatarBase64 != "" {
		params["avatar_base64"] = avatarBase64
	}

	var user User
	if err := c.Post(ctx, "/me", params, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
