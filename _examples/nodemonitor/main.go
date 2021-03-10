package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"
)

var (
	nodes = flag.String("nodes", "", "mixin node ids, join with ';'")
)

func main() {
	flag.Parse()

	if *nodes == "" {
		log.Println("./nodemonitor -nodes a;b;c")
		return
	}

	var (
		ctx      = context.Background()
		sleepDur = time.Millisecond
		monitor  = NewMonitor(strings.Split(*nodes, ";"))
	)

	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(sleepDur):
			if err := monitor.HealthCheck(ctx); err != nil {
				sleepDur = time.Second
				continue
			}

			sleepDur = time.Second * 30
		}
	}
}
