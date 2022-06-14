package mixin

import (
	"context"
)

const (
	RelationshipActionAdd     = "ADD"
	RelationshipActionRemove  = "Remove"
	RelationshipActionUpdate  = "UPDATE"
	RelationshipActionBlock   = "BLOCK"
	RelationshipActionUnblock = "UNBLOCK"
)

type RelationshipRequest struct {
	UserID   string `json:"user_id,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Action   string `json:"action,omitempty"`
}

func (c *Client) UpdateRelationship(ctx context.Context, req RelationshipRequest) (*User, error) {
	var resp User
	if err := c.Post(ctx, "/relationships", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) AddFriend(ctx context.Context, userID, remark string) (*User, error) {
	return c.UpdateRelationship(ctx, RelationshipRequest{
		UserID:   userID,
		FullName: remark,
		Action:   RelationshipActionAdd,
	})
}

func (c *Client) RemoveFriend(ctx context.Context, userID string) (*User, error) {
	return c.UpdateRelationship(ctx, RelationshipRequest{
		UserID: userID,
		Action: RelationshipActionRemove,
	})
}

func (c *Client) RemarkFriend(ctx context.Context, userID, remark string) (*User, error) {
	return c.UpdateRelationship(ctx, RelationshipRequest{
		UserID:   userID,
		FullName: remark,
		Action:   RelationshipActionUpdate,
	})
}

func (c *Client) BlockUser(ctx context.Context, userID string) (*User, error) {
	return c.UpdateRelationship(ctx, RelationshipRequest{
		UserID: userID,
		Action: RelationshipActionBlock,
	})
}

func (c *Client) UnblockUser(ctx context.Context, userID string) (*User, error) {
	return c.UpdateRelationship(ctx, RelationshipRequest{
		UserID: userID,
		Action: RelationshipActionUnblock,
	})
}

func (c *Client) ListBlockingUsers(ctx context.Context) ([]*User, error) {
	var users []*User
	if err := c.Get(ctx, "/blocking_users", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}
