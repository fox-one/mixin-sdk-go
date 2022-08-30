package mixin

import (
	"context"
	"fmt"
	"time"
)

const (
	ConversationCategoryContact = "CONTACT"
	ConversationCategoryGroup   = "GROUP"

	ParticipantActionCreate = "CREATE"
	ParticipantActionUpdate = "UPDATE"
	ParticipantActionAdd    = "ADD"
	ParticipantActionRemove = "REMOVE"
	ParticipantActionJoin   = "JOIN"
	ParticipantActionExit   = "EXIT"
	ParticipantActionRole   = "ROLE"

	ParticipantRoleOwner  = "OWNER"
	ParticipantRoleAdmin  = "ADMIN"
	ParticipantRoleMember = ""
)

// Participant conversation participant
type Participant struct {
	Action        string    `json:"action,omitempty"`
	Type          string    `json:"type,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	ParticipantID string    `json:"participant_id,omitempty"`
	Role          string    `json:"role,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

// Conversation conversation
type Conversation struct {
	ConversationID string `json:"conversation_id,omitempty"`
	CreatorID      string `json:"creator_id,omitempty"`
	Category       string `json:"category,omitempty"`
	Name           string `json:"name,omitempty"`
	IconURL        string `json:"icon_url,omitempty"`
	Announcement   string `json:"announcement,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	CodeID         string `json:"code_id,omitempty"`
	CodeURL        string `json:"code_url,omitempty"`

	Participants []*Participant `json:"participants,omitempty"`
}

type CreateConversationInput struct {
	Category       string         `json:"category,omitempty"`
	ConversationID string         `json:"conversation_id,omitempty"`
	Name           string         `json:"name,omitempty"`
	Participants   []*Participant `json:"participants,omitempty"`
}

type ConversationUpdate struct {
	Name         string `json:"name,omitempty"`
	Announcement string `json:"announcement,omitempty"`
}

// CreateConversation crate conversation
func (c *Client) CreateConversation(ctx context.Context, input *CreateConversationInput) (*Conversation, error) {
	var conversation Conversation
	if err := c.Post(ctx, "/conversations", input, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

// UpdateConversation update conversation
func (c *Client) UpdateConversation(ctx context.Context, conversationID string, input ConversationUpdate) (*Conversation, error) {
	var conversation Conversation
	if err := c.Post(ctx, fmt.Sprintf("/conversations/%s", conversationID), input, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

// CreateContactConversation create a conversation with a mixin messenger user
func (c *Client) CreateContactConversation(ctx context.Context, userID string) (*Conversation, error) {
	return c.CreateConversation(ctx, &CreateConversationInput{
		Category:       ConversationCategoryContact,
		ConversationID: UniqueConversationID(c.ClientID, userID),
		Participants:   []*Participant{{UserID: userID}},
	})
}

// CreateGroupConversation create a group in mixin messenger with given participants
func (c *Client) CreateGroupConversation(ctx context.Context, conversationID, name string, participants []*Participant) (*Conversation, error) {
	return c.CreateConversation(ctx, &CreateConversationInput{
		Category:       ConversationCategoryGroup,
		ConversationID: conversationID,
		Name:           name,
		Participants:   participants,
	})
}

// ReadConversation read conversation
func (c *Client) ReadConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	uri := fmt.Sprintf("/conversations/%s", conversationID)

	var conversation Conversation
	if err := c.Get(ctx, uri, nil, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

// Update conversation announcement
func (c *Client) UpdateConversationAnnouncement(ctx context.Context, conversationID, announcement string) (*Conversation, error) {
	return c.UpdateConversation(ctx, conversationID, ConversationUpdate{Announcement: announcement})
}

func (c *Client) ManageConversation(ctx context.Context, conversationID, action string, participants []*Participant) (*Conversation, error) {
	uri := fmt.Sprintf("/conversations/%s/participants/%s", conversationID, action)

	var conversation Conversation
	if err := c.Post(ctx, uri, participants, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (c *Client) AddParticipants(ctx context.Context, conversationID string, users ...string) (*Conversation, error) {
	var participants []*Participant
	for _, user := range users {
		participants = append(participants, &Participant{UserID: user})
	}

	return c.ManageConversation(ctx, conversationID, ParticipantActionAdd, participants)
}

func (c *Client) RemoveParticipants(ctx context.Context, conversationID string, users ...string) (*Conversation, error) {
	var participants []*Participant
	for _, user := range users {
		participants = append(participants, &Participant{UserID: user})
	}

	return c.ManageConversation(ctx, conversationID, ParticipantActionRemove, participants)
}

func (c *Client) AdminParticipants(ctx context.Context, conversationID string, users ...string) (*Conversation, error) {
	var participants []*Participant
	for _, user := range users {
		participants = append(participants, &Participant{
			UserID: user,
			Role:   ParticipantRoleAdmin,
		})
	}

	return c.ManageConversation(ctx, conversationID, ParticipantActionRole, participants)
}

func (c *Client) RotateConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	uri := fmt.Sprintf("/conversations/%s/rotate", conversationID)

	var conversation Conversation
	if err := c.Post(ctx, uri, nil, &conversation); err != nil {
		return nil, err
	}

	return &conversation, nil
}
