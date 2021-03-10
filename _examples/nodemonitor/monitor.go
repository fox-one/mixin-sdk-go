package main

import (
	"context"
	"errors"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/sirupsen/logrus"
)

type (
	Work struct {
		timestamp int64
		warnedAt  int64
		work      uint64
	}

	Monitor struct {
		works map[string]*Work
	}
)

func NewMonitor(nodes []string) *Monitor {
	works := make(map[string]*Work, len(nodes))
	for _, n := range nodes {
		works[n] = &Work{}
	}
	return &Monitor{
		works: works,
	}
}

func (m *Monitor) HealthCheck(ctx context.Context) error {
	info, err := mixin.ReadConsensusInfo(ctx)
	if err != nil {
		logrus.WithError(err).Info("ReadConsensusInfo failed")
		return err
	}

	if info.Timestamp.Before(time.Now().Add(-60 * time.Second)) {
		logrus.WithFields(logrus.Fields{
			"timestamp": info.Timestamp,
			"node":      info.Node,
		}).Info("consensus info outdated")
		return errors.New("consensus info outdated")
	}

	for _, node := range info.Graph.Consensus {
		n := node.Node.String()
		w, ok := m.works[n]
		if !ok {
			continue
		}

		cache, ok := info.Graph.Cache[n]
		if !ok || cache.Timestamp < w.timestamp {
			continue
		}

		v := node.Works[0]*12 + node.Works[1]*10
		now := time.Now().UnixNano()
		log := logrus.WithFields(logrus.Fields{
			"node":      n,
			"works":     w.work,
			"new_works": v - w.work,
			"time_diff": time.Duration(now - w.timestamp),
		})

		if v <= w.work {
			if now-w.warnedAt > int64(300*time.Second) {
				log.Info("not worked")
				w.warnedAt = now
			}
			continue
		}

		if w.warnedAt > 0 {
			log.Info("back to work")
		}
		w.warnedAt = 0
		m.works[n].work = v
		m.works[n].timestamp = cache.Timestamp
	}
	return nil
}
