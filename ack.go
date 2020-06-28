package mixin

import "context"

type AcknowledgementRequest struct {
	MessageID string `json:"message_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

func (c *Client) SendAcknowledgements(ctx context.Context, requests []*AcknowledgementRequest) error {
	if len(requests) == 0 {
		return nil
	}

	return c.Post(ctx, "/acknowledgements", requests, nil)
}

func (c *Client) SendAcknowledgement(ctx context.Context, request *AcknowledgementRequest) error {
	return c.SendAcknowledgements(ctx, []*AcknowledgementRequest{request})
}
