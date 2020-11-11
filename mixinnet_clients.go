package mixin

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	mixinnetHosts = []string{
		"http://node-42.f1ex.io:8239",
		"http://node-candy.f1ex.io:8239",
		"http://node-box.f1ex.io:8239",
		"http://node-box-2.f1ex.io:8239",
	}

	mixinNetClients = map[string]*resty.Client{}
)

func UseMixinNetHosts(hosts []string) {
	if len(hosts) == 0 {
		panic("empty mixin net host")
	}
	mixinnetHosts = hosts
	mixinNetClients = map[string]*resty.Client{}
}

func MixinNetClientFromContext(ctx context.Context) *resty.Client {
	var host string
	if v := ctx.Value(mixinnetHostKey); v != nil {
		if h, ok := v.(string); ok && h != "" {
			host = h
		}
	}
	if host == "" {
		host = RandomMixinNetHost()
	}

	if client, ok := mixinNetClients[host]; ok {
		return client
	}

	client := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHostURL(host).
		SetTimeout(10 * time.Second)

	mixinNetClients[host] = client
	return client
}

func WithMixinNetHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, mixinnetHostKey, host)
}

func RandomMixinNetHost() string {
	return mixinnetHosts[rand.Int()%len(mixinnetHosts)]
}
