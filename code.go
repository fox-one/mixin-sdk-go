package mixin

import (
	"context"
	"encoding/json"
	"fmt"
)

type CodeType string

const (
	TypeUser          CodeType = "user"
	TypeConversation  CodeType = "conversation"
	TypePayment       CodeType = "payment"
	TypeMultisig      CodeType = "multisig_request"
	TypeCollectible   CodeType = "non_fungible_request"
	TypeAuthorization CodeType = "authorization"
)

type Code struct {
	Type    CodeType `json:"type"`
	RawData json.RawMessage
}

func (c *Code) User() *User {
	if c.Type != TypeUser {
		return nil
	}
	var user User
	if err := json.Unmarshal(c.RawData, &user); err != nil {
		return nil
	}
	return &user
}

func (c *Code) Conversation() *Conversation {
	if c.Type != TypeConversation {
		return nil
	}
	var conversation Conversation
	if err := json.Unmarshal(c.RawData, &conversation); err != nil {
		return nil
	}
	return &conversation
}

func (c *Code) Payment() *Payment {
	if c.Type != TypePayment {
		return nil
	}
	var payment Payment
	if err := json.Unmarshal(c.RawData, &payment); err != nil {
		return nil
	}
	return &payment
}

func (c *Code) Multisig() *MultisigRequest {
	if c.Type != TypeMultisig {
		return nil
	}
	var multisig MultisigRequest
	if err := json.Unmarshal(c.RawData, &multisig); err != nil {
		return nil
	}
	return &multisig
}

func (c *Code) Collectible() *CollectibleRequest {
	if c.Type != TypeCollectible {
		return nil
	}
	var collectible CollectibleRequest
	if err := json.Unmarshal(c.RawData, &collectible); err != nil {
		return nil
	}
	return &collectible
}

func (c *Code) Authorization() *Authorization {
	if c.Type != TypeAuthorization {
		return nil
	}
	var authorization Authorization
	if err := json.Unmarshal(c.RawData, &authorization); err != nil {
		return nil
	}
	return &authorization
}

func (c *Client) GetCode(ctx context.Context, codeString string) (*Code, error) {
	uri := fmt.Sprintf("/codes/%s", codeString)
	var data json.RawMessage
	if err := c.Get(ctx, uri, nil, &data); err != nil {
		return nil, err
	}

	var code Code
	if err := json.Unmarshal(data, &code); err != nil {
		return nil, err
	}
	code.RawData = data
	return &code, nil
}
