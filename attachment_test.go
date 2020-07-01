package mixin

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadAttachment(t *testing.T) {
	s := newKeystoreFromTestData(t)
	c, err := NewFromKeystore(s)
	require.Nil(t, err, "init client from keystore")

	ctx := context.Background()
	attachment, err := c.CreateAttachment(ctx)
	require.Nil(t, err, "create attachment")

	data := make([]byte, 0, 64)
	_, _ = rand.Read(data)

	err = UploadAttachment(ctx, attachment, data)
	assert.Nil(t, err, "upload attachment")
}
