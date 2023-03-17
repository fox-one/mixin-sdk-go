package mixin

import (
	"context"
	"time"
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
	if key, err := KeyFromString(pin); err == nil {
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
