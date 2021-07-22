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
		RedirectURI      string    `json:"redirect_uri,omitempty"`
		HomeURI          string    `json:"home_uri,omitempty"`
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
	var app App
	uri := fmt.Sprintf("/apps/%s", appID)
	if err := c.Get(ctx, uri, nil, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

type UpdateAppRequest struct {
	RedirectURI      string   `json:"redirect_uri,omitempty"`
	HomeURI          string   `json:"home_uri,omitempty"`
	Name             string   `json:"name,omitempty"`
	Description      string   `json:"description,omitempty"`
	IconBase64       string   `json:"icon_base64,omitempty"`
	SessionSecret    string   `json:"session_secret,omitempty"`
	Category         string   `json:"category,omitempty"`
	Capabilities     []string `json:"capabilities,omitempty"`
	ResourcePatterns []string `json:"resource_patterns,omitempty"`
}

func (c *Client) UpdateApp(ctx context.Context, appID string, req UpdateAppRequest) (*App, error) {
	var app App
	uri := fmt.Sprintf("/apps/%s", appID)
	if err := c.Post(ctx, uri, req, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *Client) ReadFavoriteApps(ctx context.Context, userID string) ([]*FavoriteApp, error) {
	uri := fmt.Sprintf("/users/%s/apps/favorite", userID)

	var apps []*FavoriteApp
	if err := c.Get(ctx, uri, nil, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

func (c *Client) FavoriteApp(ctx context.Context, appID string) (*FavoriteApp, error) {
	uri := fmt.Sprintf("/apps/%s/favorite", appID)

	var app FavoriteApp
	if err := c.Post(ctx, uri, nil, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *Client) UnfavoriteApp(ctx context.Context, appID string) error {
	uri := fmt.Sprintf("/apps/%s/unfavorite", appID)

	if err := c.Post(ctx, uri, nil, nil); err != nil {
		return err
	}

	return nil
}
