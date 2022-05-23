package mixin

import (
	"context"
	"time"
)

type CollectibleCollection struct {
	CollectionID string    `json:"collection_id,omitempty"`
	Name         string    `json:"name,omitempty"`
	Type         string    `json:"type,omitempty"`
	IconUrl      string    `json:"icon_url,omitempty"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

// ReadCollectibleCollection request collectible collection
func (c *Client) ReadCollectibleCollection(ctx context.Context, collectionID string) (*CollectibleCollection, error) {
	var collection CollectibleCollection
	if err := c.Get(ctx, "/collectibles/collections/"+collectionID, nil, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

// ReadCollectibleCollection request collectible collection with accessToken
func ReadCollectibleCollection(ctx context.Context, accessToken, collectionID string) (*CollectibleCollection, error) {
	return NewFromAccessToken(accessToken).ReadCollectibleCollection(ctx, collectionID)
}
