package mixin

import (
	"context"
	"testing"
	"time"
)

func TestClient_LoopBlaze(t *testing.T) {
	s := newKeystoreFromEnv(t)
	c, err := NewFromKeystore(s)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = c.LoopBlaze(ctx, BlazeListenFunc(func(ctx context.Context, msg *MessageView, userID string) error {
		t.Log(msg.Category, msg.Data)
		return nil
	}))

	t.Log(err)
}
