package mixin

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	mixinnetHosts = []string{
		"http://node-42.f1ex.io:8239",
		"http://mixin-node-01.b.watch:8239",
		"http://mixin-node-02.b.watch:8239",
		"http://lehigh.hotot.org:8239",
		"http://node-okashi.mixin.fan:8239",
	}

	mixinNetClients sync.Map
)

func UseMixinNetHosts(hosts []string) {
	if len(hosts) == 0 {
		panic("empty mixin net host")
	}
	mixinnetHosts = hosts
	mixinNetClients = sync.Map{}
}

func MixinNetClientFromContext(ctx context.Context) *resty.Client {
	host, _ := ctx.Value(mixinnetHostKey).(string)
	if host == "" {
		host = RandomMixinNetHost()
	}

	if v, ok := mixinNetClients.Load(host); ok {
		return v.(*resty.Client)
	}

	client := resty.New().
		SetHeader("Content-Type", "application/json").
		SetBaseURL(host).
		SetTimeout(10 * time.Second)

	mixinNetClients.Store(host, client)
	return client
}

func WithMixinNetHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, mixinnetHostKey, host)
}

func RandomMixinNetHost() string {
	return mixinnetHosts[rand.Int()%len(mixinnetHosts)]
}
