package mixin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_urlScheme_Transfer(t *testing.T) {
	userID := newUUID()
	url := URL.Transfer(userID)
	assert.Equal(t, "mixin://transfer/"+userID, url)
}
