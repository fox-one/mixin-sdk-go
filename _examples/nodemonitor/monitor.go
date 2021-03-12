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

		time     time.Time
		work     uint64
		topology uint64
		warnedAt int64
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

			sleepDur = time.Second * 120
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

		var (
			t                  time.Time
			cacheSnapshotCount int
			work               = node.Works[0]*12 + node.Works[1]*10
			now                = time.Now()
		)

		if cache, ok := info.Graph.Cache[node.Node.String()]; ok && len(cache.Snapshots) > 0 {
			t = time.Unix(0, cache.Timestamp)
			cacheSnapshotCount = len(cache.Snapshots)
		} else if final, ok := info.Graph.Final[node.Node.String()]; ok {
			t = time.Unix(0, final.End)
		}

		log := log.WithFields(logrus.Fields{
			"node":            info.Node,
			"version":         info.Version,
			"topology":        info.Graph.Topology,
			"topology.pre":    m.topology,
			"cache.snapshots": cacheSnapshotCount,
			"works":           work,
			"work.pre":        m.work,
			"works.diff":      work - m.work,
			"info.time":       info.Timestamp,
			"time":            t,
			"time.pre":        m.time,
		})

		if !t.After(m.time) {
			if now.UnixNano()-m.warnedAt > int64(600*time.Second) {
				log.Infof("(%s) not working for %v", m.host, info.Timestamp.Sub(m.time))
				m.warnedAt = now.UnixNano()
			}
			continue
		}

		if m.warnedAt > 0 {
			log.Infof("(%s) back to work after %v", m.host, info.Timestamp.Sub(m.time))
		}
		m.warnedAt = 0
		m.work = work
		m.time = t
		m.topology = info.Graph.Topology
	}
	return nil
}
