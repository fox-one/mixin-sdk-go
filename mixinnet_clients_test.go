package mixin

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestMixinNetClientFromContext(t *testing.T) {
	ctx := context.Background()
	UseMixinNetHosts(mixinnetHosts)

	var g errgroup.Group
	for i := 0; i < 1000; i++ {
		g.Go(func() error {
			c1 := MixinNetClientFromContext(ctx)
			if c1 == nil {
				t.Error("client is nil")
			}

			ctx := WithMixinNetHost(ctx, c1.BaseURL)
			c2 := MixinNetClientFromContext(ctx)
			if c1.BaseURL != c2.BaseURL {
				t.Error("client is not same")
			}

			return nil
		})
	}

	_ = g.Wait()
}
