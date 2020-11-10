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

	defaultMixinNetHost = 0
	mixinNetClients     = map[string]*resty.Client{}
)

func UseMixinNetHosts(hosts []string, defaultHost int) {
	if len(hosts) == 0 || defaultHost < 0 || defaultHost > len(hosts) {
		return
	}
	mixinnetHosts = hosts
	defaultMixinNetHost = defaultHost
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
		host = mixinnetHosts[defaultMixinNetHost]
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

func WithMixinNetClient(ctx context.Context, index int) context.Context {
	if index >= 0 && index < len(mixinnetHosts) {
		ctx = context.WithValue(ctx, mixinnetHostKey, mixinnetHosts[index])
	}
	return ctx
}

func WithRandomMixinNetClient(ctx context.Context) context.Context {
	return WithMixinNetClient(ctx, rand.Int()%len(mixinnetHosts))
}
