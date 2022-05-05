package mixin

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversation(t *testing.T) {
	ctx := context.Background()
	store := newKeystoreFromEnv(t)

	c, err := NewFromKeystore(store)
	require.Nil(t, err, "init client")

	me, err := c.UserMe(ctx)
	require.Nil(t, err, "UserMe")

	id := newUUID()

	t.Run("create group conversation", func(t *testing.T) {
		conversation, err := c.CreateGroupConversation(ctx, id, "group", []*Participant{
			{
				UserID: me.App.CreatorID,
			},
		})

		require.Nil(t, err, "create conversation")
		assert.Equal(t, id, conversation.ConversationID, "check conversation id")
	})

	t.Run("send message", func(t *testing.T) {
		req := &MessageRequest{
			ConversationID: id,
			MessageID:      newUUID(),
			Category:       MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString([]byte("hello mixin-sdk-go")),
		}

		require.Nil(t, c.SendMessage(ctx, req), "send message")
	})

	t.Run("send messages", func(t *testing.T) {
		var requests []*MessageRequest

		requests = append(requests, &MessageRequest{
			ConversationID: id,
			MessageID:      newUUID(),
			RecipientID:    me.App.CreatorID,
			Category:       MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString([]byte("1")),
		})

		requests = append(requests, &MessageRequest{
			ConversationID: id,
			MessageID:      newUUID(),
			RecipientID:    me.App.CreatorID,
			Category:       MessageCategoryPlainText,
			Data:           base64.StdEncoding.EncodeToString([]byte("2")),
		})

		require.Nil(t, c.SendMessages(ctx, requests), "send messages")
	})

	t.Run("admin participant", func(t *testing.T) {
		_, err := c.AdminParticipants(ctx, id, me.App.CreatorID)
		require.Nil(t, err, "admin participant")
	})

	t.Run("remove participant", func(t *testing.T) {
		_, err := c.RemoveParticipants(ctx, id, me.App.CreatorID)
		require.Nil(t, err, "remove participant")
	})

	t.Run("add participant", func(t *testing.T) {
		_, err := c.AddParticipants(ctx, id, me.App.CreatorID)
		require.Nil(t, err, "add participant")
	})

	t.Run("rotate conversation", func(t *testing.T) {
		conversation, err := c.ReadConversation(ctx, id)
		require.Nil(t, err, "read conversation")

		rotated, err := c.RotateConversation(ctx, id)
		require.Nil(t, err, "rotate conversation")

		assert.NotEqual(t, conversation.CodeURL, rotated.CodeURL, "code url should changed")
	})

	t.Run("update announcement", func(t *testing.T) {
		conversation, err := c.ReadConversation(ctx, id)
		require.Nil(t, err, "read conversation")
		newAnnouncement := conversation.Announcement + " new"

		updated, err := c.UpdateConversationAnnouncement(ctx, id, "test")
		require.Nil(t, err, "update conversation")

		assert.Equal(t, newAnnouncement, updated.Announcement, "announcement should changed")
	})
}
