package mixin

import (
	"errors"
	"fmt"
)

// mixin error codes https://developers.mixin.one/api/alpha-mixin-network/errors/
const (
	// RequestFailed request failed
	RequestFailed = 1000000

	// InvalidTraceID invalid trace
	InvalidTraceID = 20125
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
