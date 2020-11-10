package mixin

import (
	"context"
)

type contextKey int

const (
	_ contextKey = iota
	signerKey
	verifierKey
	requestIdKey
	mixinnetHostKey
)

func WithSigner(ctx context.Context, s Signer) context.Context {
	return context.WithValue(ctx, signerKey, s)
}

func WithVerifier(ctx context.Context, v Verifier) context.Context {
	return context.WithValue(ctx, verifierKey, v)
}

// WithRequestID bind request id to context
// request id must be uuid
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIdKey, requestID)
}

var newRequestID = newUUID

func RequestIdFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIdKey).(string); ok {
		return v
	}

	return newRequestID()
}
