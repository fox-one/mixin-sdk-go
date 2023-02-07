package mixin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	xRequestID           = http.CanonicalHeaderKey("x-request-id")
	xIntegrityToken      = http.CanonicalHeaderKey("x-integrity-token")
	xForceAuthentication = http.CanonicalHeaderKey("x-force-authentication")

	ErrResponseVerifyFailed = errors.New("response verify failed")
)

var httpClient = resty.New().
	SetHeader("Content-Type", "application/json").
	SetBaseURL(DefaultApiHost).
	SetTimeout(10 * time.Second).
	SetPreRequestHook(func(c *resty.Client, r *http.Request) error {
		ctx := r.Context()
		requestID := r.Header.Get(xRequestID)
		if requestID == "" {
			requestID = RequestIdFromContext(ctx)
			r.Header.Set(xRequestID, requestID)
		}

		if s, ok := ctx.Value(signerKey).(Signer); ok {
			token := s.SignToken(SignRequest(r), requestID, time.Minute)
			r.Header.Set("Authorization", "Bearer "+token)
			r.Header.Set(xForceAuthentication, "true")
		}

		return nil
	}).
	OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		if r.IsError() {
			return nil
		}

		if err := checkResponseRequestID(r); err != nil {
			return err
		}

		if v, ok := r.Request.Context().Value(verifierKey).(Verifier); ok {
			if err := v.Verify(r); err != nil {
				return ErrResponseVerifyFailed
			}
		}

		return nil
	})

func GetClient() *http.Client {
	return httpClient.GetClient()
}

func GetRestyClient() *resty.Client {
	return httpClient
}

func checkResponseRequestID(r *resty.Response) error {
	expect := r.Request.Header.Get(xRequestID)
	got := r.Header().Get(xRequestID)
	if expect != got {
		return fmt.Errorf("%s mismatch, expect %q but got %q", xRequestID, expect, got)
	}

	return nil
}

func Request(ctx context.Context) *resty.Request {
	return httpClient.R().SetContext(ctx)
}

func DecodeResponse(resp *resty.Response) ([]byte, error) {
	var body struct {
		Error *Error          `json:"error,omitempty"`
		Data  json.RawMessage `json:"data,omitempty"`
	}

	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		if resp.IsError() {
			return nil, createError(resp.StatusCode(), resp.StatusCode(), resp.Status())
		}

		return nil, createError(resp.StatusCode(), resp.StatusCode(), err.Error())
	}

	if body.Error != nil && body.Error.Code > 0 {
		return nil, body.Error
	}
	return body.Data, nil
}

func UnmarshalResponse(resp *resty.Response, v interface{}) (err error) {
	if requestID := extractRequestID(resp); requestID != "" {
		defer bindRequestID(&err, requestID)
	}

	data, err := DecodeResponse(resp)
	if err != nil {
		return err
	}

	if v != nil {
		return json.Unmarshal(data, v)
	}

	return nil
}

func extractRequestID(r *resty.Response) string {
	if r != nil {
		return r.Request.Header.Get(xRequestID)
	}

	return ""
}

func bindRequestID(errp *error, id string) {
	if err := *errp; err != nil {
		*errp = WrapErrWithRequestID(err, id)
	}
}
