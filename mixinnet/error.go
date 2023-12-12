package mixinnet

import (
	"errors"
	"fmt"
	"strings"
)

// mixin error codes https://developers.mixin.one/api/alpha-mixin-network/errors/
const (
	InvalidOutputKey = 2000001
	InputLocked      = 2000002
	InvalidSignature = 2000003
)

type Error struct {
	Status      int                    `json:"status"`
	Code        int                    `json:"code"`
	Description string                 `json:"description"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
}

func (e *Error) Error() string {
	s := fmt.Sprintf("[%d/%d] %s", e.Status, e.Code, e.Description)
	for k, v := range e.Extra {
		s += fmt.Sprintf(" %v=%v", k, v)
	}

	if e.RequestID != "" {
		s += fmt.Sprintf(" id=%s", e.RequestID)
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

func parseError(errMsg string) error {
	if strings.HasPrefix(errMsg, "invalid output key ") {
		return createError(202, InvalidOutputKey, errMsg)
	}

	if strings.HasPrefix(errMsg, "input locked for transaction ") {
		return createError(202, InputLocked, errMsg)
	}

	if strings.HasPrefix(errMsg, "invalid tx signature number ") ||
		strings.HasPrefix(errMsg, "invalid signature keys ") {
		return createError(202, InvalidSignature, errMsg)
	}

	return createError(202, 202, errMsg)
}
