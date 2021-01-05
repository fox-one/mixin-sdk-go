package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestIDInError(t *testing.T) {
	ctx := context.Background()
	client := NewFromAccessToken("anonymous")
	_, err := client.UserMe(ctx)
	require.NotNil(t, err, "401 unauthorised")

	if e, ok := err.(*Error); ok {
		require.NotEmpty(t, e.RequestID, "request id should in error")
	}
}
