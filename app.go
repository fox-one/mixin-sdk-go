package mixin

import (
	"context"
	"fmt"
	"time"
)

type (
	App struct {
		UpdatedAt        time.Time `json:"updated_at,omitempty"`
		AppID            string    `json:"app_id,omitempty"`
		AppNumber        string    `json:"app_number,omitempty"`
		RedirectURL      string    `json:"redirect_url,omitempty"`
		HomeURL          string    `json:"home_url,omitempty"`
		Name             string    `json:"name,omitempty"`
		IconURL          string    `json:"icon_url,omitempty"`
		Description      string    `json:"description,omitempty"`
		Capabilities     []string  `json:"capabilities,omitempty"`
		ResourcePatterns []string  `json:"resource_patterns,omitempty"`
		Category         string    `json:"category,omitempty"`
		CreatorID        string    `json:"creator_id,omitempty"`
		AppSecret        string    `json:"app_secret,omitempty"`
	}

	FavoriteApp struct {
		UserID    string    `json:"user_id,omitempty"`
		AppID     string    `json:"app_id,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
	}
)

func (c *Client) ReadApp(ctx context.Context, appID string) (*App, error) {
	user, err := c.ReadUser(ctx, appID)
	if err != nil {
		return nil, err
	}

	return user.App, nil
}

func (c *Client) ReadFavoriteApps(ctx context.Context, userID string) ([]*FavoriteApp, error) {
	uri := fmt.Sprintf("/users/%s/app/favorite", userID)

	var apps []*FavoriteApp
	if err := c.Get(ctx, uri, nil, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}
