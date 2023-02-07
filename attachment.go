package mixin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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

var uploadClient = &http.Client{}

func UploadAttachmentTo(ctx context.Context, uploadURL string, file []byte) error {
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(file))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("x-amz-acl", "public-read")
	req.Header.Add("Content-Length", strconv.Itoa(len(file)))

	resp, err := uploadClient.Do(req)
	if resp != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	return nil
}

func UploadAttachment(ctx context.Context, attachment *Attachment, file []byte) error {
	return UploadAttachmentTo(ctx, attachment.UploadURL, file)
}
