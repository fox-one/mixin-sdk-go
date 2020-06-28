package mixin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

type Attachment struct {
	AttachmentID string `json:"attachment_id"`
	UploadURL    string `json:"upload_url"`
	ViewURL      string `json:"view_url"`
}

func (c *Client) CreateAttachment(ctx context.Context) (*Attachment, error) {
	var attachment Attachment
	if err := c.Post(ctx, "/attachments", nil, &attachment); err != nil {
		return nil, err
	}

	return &attachment, nil
}

func (c *Client) ShowAttachment(ctx context.Context, id string) (*Attachment, error) {
	uri := fmt.Sprintf("/attachments/%s", id)

	var attachment Attachment
	if err := c.Get(ctx, uri, nil, &attachment); err != nil {
		return nil, err
	}

	return &attachment, nil
}
func UploadAttachment(ctx context.Context, attachment *Attachment, file []byte) error {
	resp, err := Request(ctx).SetBody(file).
		SetHeader("Content-Type", "application/octet-stream").
		SetHeader("x-amz-acl", "public-read").
		SetHeader("Content-Length", strconv.Itoa(len(file))).
		Put(attachment.UploadURL)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return errors.New(resp.Status())
	}

	return nil
}
