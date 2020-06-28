package mixin

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Signer interface {
	SignToken(signature, requestID string, exp time.Duration) string
	EncryptPin(pin string) string
}

type Verifier interface {
	Verify(resp *resty.Response) error
}

type nopVerifier struct{}

func (nopVerifier) Verify(_ *resty.Response) error {
	return nil
}

func NopVerifier() Verifier {
	return nopVerifier{}
}
