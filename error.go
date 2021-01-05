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

// errWithRequestID wrap err with request id
type errWithRequestID struct {
	err       error
	requestID string
}

func (e *errWithRequestID) Unwrap() error {
	return e.err
}

func (e *errWithRequestID) Error() string {
	return fmt.Sprintf("%v id=%s", e.err, e.requestID)
}

func WrapErrWithRequestID(err error, id string) error {
	if e, ok := err.(*Error); ok {
		e.RequestID = id
		return e
	}

	return &errWithRequestID{err: err, requestID: id}
}
