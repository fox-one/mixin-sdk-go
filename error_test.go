package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsErrorCodes(t *testing.T) {
	_, err := UserMe(context.TODO(), "invalid token")
	assert.True(t, IsErrorCodes(err, Unauthorized), "error should be %v", Unauthorized)
}
