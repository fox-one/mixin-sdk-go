package mixin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/gorilla/websocket"
)

type Authorization struct {
	CreatedAt         time.Time `json:"created_at"`
	AccessedAt        time.Time `json:"accessed_at"`
	AuthorizationID   string    `json:"authorization_id"`
	AuthorizationCode string    `json:"authorization_code"`
	Scopes            []string  `json:"scopes"`
	CodeID            string    `json:"code_id"`
	App               App       `json:"app"`
	User              User      `json:"user"`
}

func (c *Client) Authorize(ctx context.Context, authorizationID string, scopes []string, pin string) (*Authorization, error) {
	if key, err := mixinnet.KeyFromString(pin); err == nil {
		pin = c.EncryptTipPin(
			key,
			TIPOAuthApprove,
			authorizationID,
		)
	}

	body := map[string]interface{}{
		"authorization_id": authorizationID,
		"scopes":           scopes,
		"pin_base64":       c.EncryptPin(pin),
	}

	var authorization Authorization
	if err := c.Post(ctx, "/oauth/authorize", body, &authorization); err != nil {
		return nil, err
	}

	return &authorization, nil
}

func RequestAuthorization(ctx context.Context, clientID string, scopes []string, challenge string) (*Authorization, error) {
	dialer := &websocket.Dialer{
		Subprotocols:   []string{"Mixin-OAuth-1"},
		ReadBufferSize: 1024,
	}

	conn, _, err := dialer.Dial(blazeURL, nil)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err := writeMessage(conn, "REFRESH_OAUTH_CODE", map[string]interface{}{
		"client_id":      clientID,
		"scope":          strings.Join(scopes, " "),
		"code_challenge": challenge,
	}); err != nil {
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	_, r, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	var msg BlazeMessage
	if err := parseBlazeMessage(r, &msg); err != nil {
		return nil, err
	}

	if msg.Error != nil {
		return nil, msg.Error
	}

	var auth Authorization
	if err := json.Unmarshal(msg.Data, &auth); err != nil {
		return nil, err
	}

	return &auth, nil
}

func CodeChallenge(b []byte) (verifier, challange string) {
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.New()
	h.Write(b)
	challange = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return
}

func RandomCodeChallenge() (verifier, challange string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return CodeChallenge(b)
}
