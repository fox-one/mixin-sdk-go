package mixin

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadAttachment(t *testing.T) {
	store := newKeystoreFromEnv(t)
	c, err := NewFromKeystore(&store.Keystore)
	require.Nil(t, err, "init client from keystore")

	ctx := context.Background()
	attachment, err := c.CreateAttachment(ctx)
	require.Nil(t, err, "create attachment")

	data := make([]byte, 128)
	_, _ = rand.Read(data)

	err = UploadAttachment(ctx, attachment, data)
	assert.Nil(t, err, "upload attachment")
}
