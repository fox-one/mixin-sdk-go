package main

import (
	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/sirupsen/logrus"
)

type (
	Monitor struct {
		host string

		timestamp int64
		work      uint64
		topology  uint64
		warnedAt  int64
	}
)

func NewMonitor(host string) *Monitor {
	return &Monitor{
		host: host,
	}
}

func (m *Monitor) LoopHealthCheck(ctx context.Context) error {
	ctx = mixin.WithMixinNetHost(ctx, m.host)
	sleepDur := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(sleepDur):
			if err := m.healthCheck(ctx); err != nil {
				sleepDur = time.Second
				continue
			}

			sleepDur = time.Second * 30
		}
	}
}

func (m *Monitor) healthCheck(ctx context.Context) error {
	log := logrus.WithFields(logrus.Fields{
		"host": m.host,
	})

	info, err := mixin.ReadConsensusInfo(ctx)
	if err != nil {
		log.WithError(err).Info("ReadConsensusInfo failed")
		return err
	}

	for _, node := range info.Graph.Consensus {
		if node.Node != info.Node {
			continue
		}

		cache, ok := info.Graph.Cache[node.Node.String()]
		if !ok {
			continue
		}

		work := node.Works[0]*12 + node.Works[1]*10
		now := time.Now().UnixNano()
		log := log.WithFields(logrus.Fields{
			"node":           info.Node,
			"version":        info.Version,
			"topology":       info.Graph.Topology,
			"topology.pre":   m.topology,
			"works":          work,
			"work.pre":       m.work,
			"works.diff":     work - m.work,
			"cache.time":     cache.Timestamp,
			"cache.time.pre": m.timestamp,
			"duration":       time.Duration(now - m.timestamp),
		})

		if work == m.work {
			if now-m.warnedAt > int64(300*time.Second) {
				log.Info("not worked")
				m.warnedAt = now
			}
			continue
		}

		if m.warnedAt > 0 {
			log.Info("back to work")
		}
		m.warnedAt = 0
		m.work = work
		m.timestamp = cache.Timestamp
		m.topology = info.Graph.Topology
	}
	return nil
}
