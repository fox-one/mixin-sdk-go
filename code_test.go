package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCode(t *testing.T) {
	ctx := context.Background()
	store := newKeystoreFromEnv(t)

	c, err := NewFromKeystore(&store.Keystore)
	require.Nil(t, err, "init client")

	code, err := c.GetCode(ctx, "c76310d8-c563-499e-9866-c61ae2cbee11")
	require.Nil(t, err, "get code")
	require.True(t, code.Type == TypePayment)
	payment := code.Payment()
	require.True(t, payment.Amount == "1")

	code, err = c.GetCode(ctx, "d4b174c2-2691-4289-b4a0-2d0f9ec43618")
	require.Nil(t, err, "get code")
	require.True(t, code.Type == TypePayment)
	payment = code.Payment()
	require.True(t, payment.Amount == "1")

	code, err = c.GetCode(ctx, "e0a53283-da39-438c-ba94-8d77071e9860")
	require.Nil(t, err, "get code")
	require.True(t, code.Type == TypeConversation)
	conversation := code.Conversation()
	require.True(t, conversation.Category == "GROUP")
}
