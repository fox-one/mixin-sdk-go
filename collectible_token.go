package mixin

import (
	"context"
	"time"
)

type CollectibleTokenMeta struct {
	Group       string `json:"group,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	MediaURL    string `json:"media_url,omitempty"`
	Mime        string `json:"mime,omitempty"`
	Hash        Hash   `json:"hash,omitempty"`
}

type CollectibleToken struct {
	Type         string               `json:"type,omitempty"`
	CreatedAt    time.Time            `json:"created_at,omitempty"`
	CollectionID string               `json:"collection_id,omitempty"`
	TokenID      string               `json:"token_id,omitempty"`
	Group        string               `json:"group,omitempty"`
	Token        string               `json:"token,omitempty"`
	MixinID      Hash                 `json:"mixin_id,omitempty"`
	NFO          TransactionExtra     `json:"nfo,omitempty"`
	Meta         CollectibleTokenMeta `json:"meta,omitempty"`
}

// ReadCollectiblesToken return the detail of CollectibleToken
func (c *Client) ReadCollectiblesToken(ctx context.Context, id string) (*CollectibleToken, error) {
	var token CollectibleToken
	if err := c.Get(ctx, "/collectibles/tokens/"+id, nil, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// ReadCollectiblesToken request with access token and returns the detail of CollectibleToken
func ReadCollectiblesToken(ctx context.Context, accessToken, tokenID string) (*CollectibleToken, error) {
	return NewFromAccessToken(accessToken).ReadCollectiblesToken(ctx, tokenID)
}
