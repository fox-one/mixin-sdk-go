package mixin

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

const (
	CircleActionAdd    = "ADD"
	CircleActionRemove = "REMOVE"

	CircleItemTypeUsers         = "users"
	CircleItemTypeConversations = "conversations"
)

type Circle struct {
	ID        string    `json:"circle_id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (c *Client) ListCircles(ctx context.Context) ([]*Circle, error) {
	var circles []*Circle
	if err := c.Get(ctx, "/circles", nil, &circles); err != nil {
		return nil, err
	}

	return circles, nil
}

func (c *Client) ReadCircle(ctx context.Context, circleID string) (*Circle, error) {
	var circle Circle
	if err := c.Get(ctx, "/circles/"+circleID, nil, &circle); err != nil {
		return nil, err
	}

	return &circle, nil
}

type CreateCircleParams struct {
	Name string `json:"name,omitempty"`
}

func (c *Client) CreateCircle(ctx context.Context, args CreateCircleParams) (*Circle, error) {
	var circle Circle
	if err := c.Post(ctx, "/circles", args, &circle); err != nil {
		return nil, err
	}

	return &circle, nil
}

type UpdateCircleParams struct {
	Name string `json:"name,omitempty"`
}

func (c *Client) UpdateCircle(ctx context.Context, circleID string, args UpdateCircleParams) (*Circle, error) {
	var circle Circle
	if err := c.Post(ctx, "/circles/"+circleID, args, &circle); err != nil {
		return nil, err
	}

	return &circle, nil
}

func (c *Client) DeleteCircle(ctx context.Context, circleID string) error {
	uri := fmt.Sprintf("/circles/%s/delete", circleID)
	return c.Post(ctx, uri, nil, nil)
}

type ManageCircleParams struct {
	Action   string `json:"action,omitempty"`    // ADD or REMOVE
	ItemType string `json:"item_type,omitempty"` // users or conversations
	ItemID   string `json:"item_id,omitempty"`
}

func (c *Client) ManageCircle(ctx context.Context, circleID string, args ManageCircleParams) ([]*Circle, error) {
	var circles []*Circle
	uri := fmt.Sprintf("%s/%s/circles", args.ItemType, args.ItemID)
	body := map[string]interface{}{
		"action":    args.Action,
		"circle_id": circleID,
	}

	if err := c.Post(ctx, uri, body, &circles); err != nil {
		return nil, err
	}

	return circles, nil
}

type CircleItem struct {
	CreatedAt      time.Time `json:"created_at,omitempty"`
	CircleID       string    `json:"circle_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	UserID         string    `json:"user_id,omitempty"`
}

func (c *Client) ListCircleItems(ctx context.Context, circleID string, offset time.Time, limit int) ([]*CircleItem, error) {
	var items []*CircleItem
	uri := fmt.Sprintf("/circles/%s/conversations", circleID)
	params := map[string]string{
		"offset": offset.Format(time.RFC3339Nano),
		"limit":  strconv.Itoa(limit),
	}

	if err := c.Get(ctx, uri, params, &items); err != nil {
		return nil, err
	}

	return items, nil
}
