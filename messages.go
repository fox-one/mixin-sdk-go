package mixin

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"time"
)

const (
	MessageCategoryPlainText             = "PLAIN_TEXT"
	MessageCategoryPlainPost             = "PLAIN_POST"
	MessageCategoryPlainImage            = "PLAIN_IMAGE"
	MessageCategoryPlainAudio            = "PLAIN_AUDIO"
	MessageCategoryPlainData             = "PLAIN_DATA"
	MessageCategoryPlainSticker          = "PLAIN_STICKER"
	MessageCategoryPlainLive             = "PLAIN_LIVE"
	MessageCategoryPlainVideo            = "PLAIN_VIDEO"
	MessageCategoryPlainContact          = "PLAIN_CONTACT"
	MessageCategoryPlainTranscript       = "PLAIN_TRANSCRIPT"
	MessageCategoryPlainLocation         = "PLAIN_LOCATION"
	MessageCategoryAppCard               = "APP_CARD"
	MessageCategoryAppButtonGroup        = "APP_BUTTON_GROUP"
	MessageCategoryMessageRecall         = "MESSAGE_RECALL"
	MessageCategorySystemConversation    = "SYSTEM_CONVERSATION"
	MessageCategorySystemAccountSnapshot = "SYSTEM_ACCOUNT_SNAPSHOT"

	MessageStatusSent      = "SENT"
	MessageStatusDelivered = "DELIVERED"
	MessageStatusRead      = "READ"
)

type (
	AttachmentMessageEncrypt struct {
		Key    []byte `json:"key"`
		Digest []byte `json:"digest"`
	}

	RecallMessage struct {
		MessageID string `json:"message_id"`
	}

	ImageMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		Width        int    `json:"width,omitempty"`
		Height       int    `json:"height,omitempty"`
		Size         int    `json:"size,omitempty"`
		Thumbnail    string `json:"thumbnail,omitempty"`
		*AttachmentMessageEncrypt
	}

	DataMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		Size         int    `json:"size,omitempty"`
		Name         string `json:"name,omitempty"`
		*AttachmentMessageEncrypt
	}

	StickerMessage struct {
		Name    string `json:"name,omitempty"`
		AlbumID string `json:"album_id,omitempty"`
	}

	ContactMessage struct {
		UserID string `json:"user_id,omitempty"`
	}

	TranscriptMessage struct {
		TranscriptID   string    `json:"transcript_id,omitempty"`
		MessageID      string    `json:"message_id,omitempty"`
		UserID         string    `json:"user_id,omitempty"`
		UserFullName   string    `json:"user_full_name,omitempty"`
		Category       string    `json:"category,omitempty"`
		Content        string    `json:"content,omitempty"`
		MediaURL       string    `json:"media_url,omitempty"`
		MediaName      string    `json:"media_name,omitempty"`
		MediaSize      int       `json:"media_size,omitempty"`
		MediaWidth     int       `json:"media_width,omitempty"`
		MediaHeight    int       `json:"media_height,omitempty"`
		MediaDuration  int       `json:"media_duration,omitempty"`
		MediaMimeType  string    `json:"media_mime_type,omitempty"`
		MediaStatus    string    `json:"media_status,omitempty"`
		MediaWaveform  string    `json:"media_waveform,omitempty"`
		MediaKey       string    `json:"media_key,omitempty"`
		MediaDigest    string    `json:"media_digest,omitempty"`
		MediaCreatedAt time.Time `json:"media_created_at,omitempty"`
		ThumbImage     string    `json:"thumb_image,omitempty"`
		ThumbURL       string    `json:"thumb_url,omitempty"`
		StickerID      string    `json:"sticker_id,omitempty"`
		SharedUserID   string    `json:"shared_user_id,omitempty"`
		Mentions       string    `json:"mentions,omitempty"`
		QuoteID        string    `json:"quote_id,omitempty"`
		QuoteContent   string    `json:"quote_content,omitempty"`
		Caption        string    `json:"caption,omitempty"`
		CreatedAt      time.Time `json:"created_at,omitempty"`
	}

	AppCardMessage struct {
		AppID       string `json:"app_id,omitempty"`
		IconURL     string `json:"icon_url,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Action      string `json:"action,omitempty"`
		Shareable   bool   `json:"shareable,omitempty"`
	}

	AudioMessage struct {
		AttachmentID string `json:"attachment_id,omitempty"`
		MimeType     string `json:"mime_type,omitempty"`
		WaveForm     string `json:"wave_form,omitempty"`
		Size         int    `json:"size,omitempty"`
		Duration     int    `json:"duration,omitempty"`
		*AttachmentMessageEncrypt
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
		*AttachmentMessageEncrypt
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

type (
	MessageRequest struct {
		ConversationID string `json:"conversation_id"`
		RecipientID    string `json:"recipient_id"`
		MessageID      string `json:"message_id"`
		Category       string `json:"category"`
		Data           string `json:"data,omitempty"`
		// DataBase64 is same as Data but encoded by base64.RawURLEncoding
		DataBase64       string `json:"data_base64,omitempty"`
		RepresentativeID string `json:"representative_id,omitempty"`
		QuoteMessageID   string `json:"quote_message_id,omitempty"`
		Silent           bool   `json:"silent,omitempty"`

		// encrypted messages
		Checksum          string             `json:"checksum,omitempty"`
		RecipientSessions []RecipientSession `json:"recipient_sessions,omitempty"`
	}

	RecipientSession struct {
		SessionID string `json:"session_id,omitempty"`
	}
)

func (c *Client) SendMessage(ctx context.Context, message *MessageRequest) error {
	raw, _ := json.Marshal(message)
	return c.SendRawMessage(ctx, raw)
}

func (c *Client) SendMessages(ctx context.Context, messages []*MessageRequest) error {
	raws := make([]json.RawMessage, 0, len(messages))
	for _, msg := range messages {
		b, _ := json.Marshal(msg)
		raws = append(raws, b)
	}

	return c.SendRawMessages(ctx, raws)
}

func (c *Client) SendRawMessage(ctx context.Context, message json.RawMessage) error {
	return c.Post(ctx, "/messages", message, nil)
}

func (c *Client) SendRawMessages(ctx context.Context, messages []json.RawMessage) error {
	return c.Post(ctx, "/messages", messages, nil)
}

func DecryptAttachment(data, keys, digest []byte) ([]byte, error) {
	aesKey := keys[:32]
	iv := data[:16]
	ciphertext := data[16 : len(data)-32]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	return ciphertext, nil
}
