package mixin

import (
	"container/list"
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type AckQueue struct {
	list list.List
	mux  sync.Mutex
}

func (q *AckQueue) pushBack(requests ...*AcknowledgementRequest) {
	q.mux.Lock()
	for _, req := range requests {
		q.list.PushBack(req)
	}
	q.mux.Unlock()
}

func (q *AckQueue) pushFront(requests ...*AcknowledgementRequest) {
	q.mux.Lock()
	for _, req := range requests {
		q.list.PushFront(req)
	}
	q.mux.Unlock()
}

func (q *AckQueue) pull(limit int) []*AcknowledgementRequest {
	q.mux.Lock()

	if limit > q.list.Len() {
		limit = q.list.Len()
	}

	ids := make([]*AcknowledgementRequest, 0, limit)
	for i := 0; i < limit; i++ {
		e := q.list.Front()
		ids = append(ids, e.Value.(*AcknowledgementRequest))
		q.list.Remove(e)
	}

	q.mux.Unlock()
	return ids
}

type blazeHandler struct {
	*Client
	queue AckQueue
}

func (b *blazeHandler) ack(ctx context.Context) error {
	var (
		g   errgroup.Group
		sem = semaphore.NewWeighted(5)
		dur = time.Second
	)

	for {
		select {
		case <-ctx.Done():
			return g.Wait()
		case <-time.After(dur):
			requests := b.queue.pull(ackBatch)
			if len(requests) < ackBatch {
				dur = time.Second
			} else {
				dur = 200 * time.Millisecond
			}

			if len(requests) > 0 {
				if !sem.TryAcquire(1) {
					b.queue.pushFront(requests...)
					break
				}

				g.Go(func() error {
					defer sem.Release(1)

					err := b.SendAcknowledgements(ctx, requests)
					if err != nil {
						b.queue.pushFront(requests...)
					}

					return err
				})
			}
		}
	}
}
