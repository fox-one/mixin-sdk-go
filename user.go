package mixin

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"
)

type User struct {
	UserID                   string    `json:"user_id,omitempty"`
	IdentityNumber           string    `json:"identity_number,omitempty"`
	Phone                    string    `json:"phone,omitempty"`
	FullName                 string    `json:"full_name,omitempty"`
	Biography                string    `json:"biography,omitempty"`
	AvatarURL                string    `json:"avatar_url,omitempty"`
	Relationship             string    `json:"relationship,omitempty"`
	MuteUntil                time.Time `json:"mute_until,omitempty"`
	CreatedAt                time.Time `json:"created_at,omitempty"`
	IsVerified               bool      `json:"is_verified,omitempty"`
	IsScam                   bool      `json:"is_scam,omitempty"`
	SessionID                string    `json:"session_id,omitempty"`
	PinToken                 string    `json:"pin_token,omitempty"`
	CodeID                   string    `json:"code_id,omitempty"`
	CodeURL                  string    `json:"code_url,omitempty"`
	HasPin                   bool      `json:"has_pin,omitempty"`
	DeviceStatus             string    `json:"device_status,omitempty"`
	HasEmergencyContact      bool      `json:"has_emergency_contact,omitempty"`
	ReceiveMessageSource     string    `json:"receive_message_source,omitempty"`
	AcceptConversationSource string    `json:"accept_conversation_source,omitempty"`
	AcceptSearchSource       string    `json:"accept_search_source,omitempty"`
	FiatCurrency             string    `json:"fiat_currency,omitempty"`

	App *App `json:"app,omitempty"`
}

type UserUpdate struct {
	FullName     string `json:"full_name,omitempty"`
	AvatarBase64 string `json:"avatar_base64,omitempty"`
	Biography    string `json:"biography,omitempty"`
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
	if err := c.Post(ctx, "/users/fetch", ids, &users); err != nil {
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

func (c *Client) SearchUser(ctx context.Context, identityNumberOrPhoneNumber string) (*User, error) {
	uri := fmt.Sprintf("/search/%s", identityNumberOrPhoneNumber)

	var user User
	if err := c.Get(ctx, uri, nil, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateUser(ctx context.Context, key crypto.Signer, fullname string) (*User, *Keystore, error) {
	publicKey := key.Public()
	var pub []byte
	switch k := publicKey.(type) {
	case ed25519.PublicKey:
		pub = k
	case *rsa.PublicKey:
		var err error
		pub, err = x509.MarshalPKIXPublicKey(publicKey)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("unexpected public key type: %T", key)
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

func newKeystoreFromUser(user *User, privateKey crypto.PrivateKey) *Keystore {
	var pk string
	switch k := privateKey.(type) {
	case ed25519.PrivateKey:
		pk = ed25519Encoding.EncodeToString(k)
	case *rsa.PrivateKey:
		s := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		})
		pk = string(s)
	}

	return &Keystore{
		ClientID:   user.UserID,
		SessionID:  user.SessionID,
		PinToken:   user.PinToken,
		PrivateKey: pk,
	}
}

func (c *Client) ModifyProfile(ctx context.Context, fullname, avatarBase64 string) (*User, error) {
	return c.UpdateProfile(ctx, UserUpdate{FullName: fullname, AvatarBase64: avatarBase64})
}

func (c *Client) UpdateProfile(ctx context.Context, input UserUpdate) (*User, error) {
	var user User
	if err := c.Post(ctx, "/me", input, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
