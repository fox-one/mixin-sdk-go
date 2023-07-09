package mixin

import (
	"crypto/rand"
	"testing"
)

func TestNewPublicMixinnetAddress(t *testing.T) {
	r := rand.Reader

	a := NewMixinnetAddress(r, true)
	b := NewMixinnetAddress(r, true)

	if a.PrivateViewKey.String() == b.PrivateViewKey.String() {
		t.Errorf("same PrivateViewKey generated %v, %v", a.PrivateViewKey, b.PrivateViewKey)
	}
}
