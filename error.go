package mixin

import (
	"errors"
	"fmt"
)

// mixin error codes https://developers.mixin.one/api/alpha-mixin-network/errors/
const (
	Unauthorized        = 401
	EndpointNotFound    = 404
	InsufficientBalance = 20117
	PinIncorrect        = 20119
	InsufficientFee     = 20124
	InvalidTraceID      = 20125
	InvalidReceivers    = 20150
)

type Error struct {
	Status      int                    `json:"status"`
	Code        int                    `json:"code"`
	Description string                 `json:"description"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

func (e *Error) Error() string {
	s := fmt.Sprintf("[%d/%d] %s", e.Status, e.Code, e.Description)
	for k, v := range e.Extra {
		s += fmt.Sprintf(" %v=%v", k, v)
	}

	return s
}

func IsErrorCodes(err error, codes ...int) bool {
	var e *Error
	if errors.As(err, &e) {
		for _, code := range codes {
			if e.Code == code {
				return true
			}
		}
	}

	return false
}

func createError(status, code int, description string) error {
	return &Error{
		Status:      status,
		Code:        code,
		Description: description,
	}
}
