package mixin

import (
	"context"
	"encoding/json"
)

const (
	MessageCategoryPlainText             = "PLAIN_TEXT"
	MessageCategoryPlainPost             = "PLAIN_POST"
	MessageCategoryPlainImage            = "PLAIN_IMAGE"
	MessageCategoryPlainData             = "PLAIN_DATA"
	MessageCategoryPlainSticker          = "PLAIN_STICKER"
	MessageCategoryPlainLive             = "PLAIN_LIVE"
	MessageCategoryPlainVideo            = "PLAIN_VIDEO"
	MessageCategoryPlainContact          = "PLAIN_CONTACT"
	MessageCategoryAppCard               = "APP_CARD"
	MessageCategoryAppButtonGroup        = "APP_BUTTON_GROUP"
	MessageCategoryMessageRecall         = "MESSAGE_RECALL"
	MessageCategorySystemConversation    = "SYSTEM_CONVERSATION"
	MessageCategorySystemAccountSnapshot = "SYSTEM_ACCOUNT_SNAPSHOT"

	MessageStatusSent = "SENT"
	MessageStatusRead = "READ"
)

type (
	RecallMessage struct {
		MessageID string `json:"message_id"`
	}

	ImageMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		Width        int    `json:"width,omitempty"`
		Height       int    `json:"height,omitempty"`
		Size         int    `json:"size,omitempty"`
		Thumbnail    []byte `json:"thumbnail,omitempty"`
	}

	DataMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		Size         int    `json:"size,omitempty"`
		Name         string `json:"name,omitempty"`
	}

	StickerMessage struct {
		Name    string `json:"name,omitempty"`
		AlbumID string `json:"album_id,omitempty"`
	}

	ContactMessage struct {
		UserID string `json:"user_id,omitempty"`
	}

	AppCardMessage struct {
		AppID       string `json:"app_id,omitempty"`
		IconURL     string `json:"icon_url,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Action      string `json:"action,omitempty"`
	}

	AudioMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		WaveForm     string `json:"wave_form,omitempty"`
		Size         int    `json:"size,omitempty"`
		Duration     int    `json:"duration,omitempty"`
	}

	LiveMessage struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		ThumbUrl string `json:"thumb_url"`
		URL      string `json:"url"`
	}

	VideoMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		WaveForm     string `json:"wave_form,omitempty"`
		Width        int    `json:"width,omitempty"`
		Height       int    `json:"height,omitempty"`
		Size         int    `json:"size,omitempty"`
		Duration     int    `json:"duration,omitempty"`
		Thumbnail    []byte `json:"thumbnail,omitempty"`
	}

	LocationMessage struct {
		Name      string  `json:"name,omitempty"`
		Address   string  `json:"address,omitempty"`
		Longitude float64 `json:"longitude,omitempty"`
		Latitude  float64 `json:"latitude,omitempty"`
	}

	AppButtonMessage struct {
		Label  string `json:"label,omitempty"`
		Action string `json:"action,omitempty"`
		Color  string `json:"color,omitempty"`
	}

	AppButtonGroupMessage []AppButtonMessage
)

type MessageRequest struct {
	ConversationID   string `json:"conversation_id"`
	RecipientID      string `json:"recipient_id"`
	MessageID        string `json:"message_id"`
	Category         string `json:"category"`
	Data             string `json:"data"`
	RepresentativeID string `json:"representative_id,omitempty"`
	QuoteMessageID   string `json:"quote_message_id,omitempty"`
}

func (c *Client) SendMessages(ctx context.Context, messages []*MessageRequest) error {
	switch len(messages) {
	case 0:
		return nil
	case 1:
		return c.SendMessage(ctx, messages[0])
	default:
		return c.Post(ctx, "/messages", messages, nil)
	}
}

func (c *Client) SendMessage(ctx context.Context, message *MessageRequest) error {
	return c.Post(ctx, "/messages", message, nil)
}

func (c *Client) SendRawMessages(ctx context.Context, messages []json.RawMessage) error {
	return c.Post(ctx, "/messages", messages, nil)
}
