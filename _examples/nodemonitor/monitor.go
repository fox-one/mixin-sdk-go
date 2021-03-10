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
		warnedAt int64
		work     uint64
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

		v := node.Works[0]*12 + node.Works[1]*10
		now := time.Now().UnixNano()
		log := log.WithFields(logrus.Fields{
			"node":       info.Node,
			"works":      v,
			"works_pre":  m.work,
			"works_diff": v - m.work,
			"info.time":  info.Timestamp,
			"since_now":  time.Now().Sub(m.time),
		})
		if v == m.work {
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
		m.work = v
		m.time = info.Timestamp
	}
	return nil
}
