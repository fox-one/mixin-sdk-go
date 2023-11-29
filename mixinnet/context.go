package mixinnet

import (
	"context"
	"math/rand"
)

type contextKey int

const (
	_ contextKey = iota
	hostKey
)

func (c *Client) HostFromContext(ctx context.Context) string {
	if host, ok := ctx.Value(hostKey).(string); ok {
		return host
	}
	return c.RandomHost()
}

func (c *Client) RandomHost() string {
	return c.hosts[rand.Int()%len(c.hosts)]
}

func (c *Client) WithHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, hostKey, host)
}
