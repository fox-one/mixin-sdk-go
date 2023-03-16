package mixin

import (
	"context"
	"encoding/json"
	"fmt"
)

type CodeType string

const (
	TypeUser         CodeType = "user"
	TypeConversation CodeType = "conversation"
	TypePayment      CodeType = "payment"
	TypeMultisig     CodeType = "multisig_request"
	TypeCollectible  CodeType = "non_fungible_request"
)

type Code struct {
	Type CodeType `json:"type"`
	Data interface{}
}

func (c Code) User() *User {
	user, ok := c.Data.(User)
	if !ok {
		return nil
	}
	return &user
}

func (c Code) Conversation() *Conversation {
	conversation, ok := c.Data.(Conversation)
	if !ok {
		return nil
	}
	return &conversation
}

func (c Code) Payment() *Payment {
	payment, ok := c.Data.(Payment)
	if !ok {
		return nil
	}
	return &payment
}

func (c Code) Multisig() *MultisigRequest {
	multisig, ok := c.Data.(MultisigRequest)
	if !ok {
		return nil
	}
	return &multisig
}

func (c Code) Collectible() *CollectibleRequest {
	collectible, ok := c.Data.(CollectibleRequest)
	if !ok {
		return nil
	}
	return &collectible
}

func (c *Client) GetCode(ctx context.Context, codeString string) (*Code, error) {
	uri := fmt.Sprintf("/codes/%s", codeString)
	r, err := c.Request(ctx).SetQueryParams(nil).Get(uri)
	if err != nil {
		return nil, err
	}

	data, err := DecodeResponse(r)
	if err != nil {
		return nil, err
	}

	return decodeCode(data)
}

func decodeCode(data []byte) (*Code, error) {
	var code Code
	err := json.Unmarshal(data, &code)
	if err != nil {
		return nil, err
	}

	switch code.Type {
	case TypeUser:
		var user User
		err := json.Unmarshal(data, &user)
		if err != nil {
			return nil, err
		}
		code.Data = user
	case TypeConversation:
		var conversation Conversation
		err := json.Unmarshal(data, &conversation)
		if err != nil {
			return nil, err
		}
		code.Data = conversation
	case TypePayment:
		var payment Payment
		err := json.Unmarshal(data, &payment)
		if err != nil {
			return nil, err
		}
		code.Data = payment
	case TypeMultisig:
		var multisigRequest MultisigRequest
		err := json.Unmarshal(data, &multisigRequest)
		if err != nil {
			return nil, err
		}
		code.Data = multisigRequest
	case TypeCollectible:
		var collectibleRequest CollectibleRequest
		err := json.Unmarshal(data, &collectibleRequest)
		if err != nil {
			return nil, err
		}
		code.Data = collectibleRequest
	}

	return &code, nil
}
