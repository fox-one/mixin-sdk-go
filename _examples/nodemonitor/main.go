package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	hosts = flag.String("hosts", "", "mixin node rpc hosts, join with ';'")
)

func main() {
	flag.Parse()

	if *hosts == "" {
		log.Println("./nodemonitor -hosts a;b;c")
		return
	}

	g, ctx := errgroup.WithContext(context.Background())
	for _, host := range strings.Split(*hosts, ";") {
		host := host
		if host == "" {
			continue
		}
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		g.Go(func() error {
			return NewMonitor(host).LoopHealthCheck(ctx)
		})
		time.Sleep(time.Millisecond * 100)
	}
	g.Wait()
}
