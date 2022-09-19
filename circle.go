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
	UserID    string    `json:"user_id,omitempty"`
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
	CircleID string `json:"circle_id,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (c *Client) UpdateCircle(ctx context.Context, args UpdateCircleParams) (*Circle, error) {
	var circle Circle
	body := map[string]interface{}{
		"name": args.Name,
	}

	if err := c.Post(ctx, "/circles/"+args.CircleID, body, &circle); err != nil {
		return nil, err
	}

	return &circle, nil
}

func (c *Client) DeleteCircle(ctx context.Context, circleID string) error {
	uri := fmt.Sprintf("/circles/%s/delete", circleID)
	return c.Post(ctx, uri, nil, nil)
}

type ManageCircleParams struct {
	CircleID string `json:"circle_id,omitempty"`
	Action   string `json:"action,omitempty"`    // ADD or REMOVE
	ItemType string `json:"item_type,omitempty"` // users or conversations
	ItemID   string `json:"item_id,omitempty"`   // user_id or conversation_id
}

type CircleItem struct {
	CreatedAt      time.Time `json:"created_at,omitempty"`
	CircleID       string    `json:"circle_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	UserID         string    `json:"user_id,omitempty"`
}

func (c *Client) ManageCircle(ctx context.Context, args ManageCircleParams) (*CircleItem, error) {
	var items []*CircleItem
	uri := fmt.Sprintf("%s/%s/circles", args.ItemType, args.ItemID)
	body := []interface{}{map[string]interface{}{
		"action":    args.Action,
		"circle_id": args.CircleID,
	}}

	if err := c.Post(ctx, uri, body, &items); err != nil {
		return nil, err
	}

	return items[0], nil
}

type ListCircleItemsParams struct {
	CircleID string    `json:"circle_id,omitempty"`
	Offset   time.Time `json:"offset,omitempty"`
	Limit    int       `json:"limit,omitempty"`
}

func (c *Client) ListCircleItems(ctx context.Context, args ListCircleItemsParams) ([]*CircleItem, error) {
	var items []*CircleItem
	uri := fmt.Sprintf("/circles/%s/conversations", args.CircleID)
	params := map[string]string{
		"offset": args.Offset.Format(time.RFC3339Nano),
		"limit":  strconv.Itoa(args.Limit),
	}

	if err := c.Get(ctx, uri, params, &items); err != nil {
		return nil, err
	}

	return items, nil
}
