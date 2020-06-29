package mixin

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

type blazeHandler struct {
	*Client
	conn         *websocket.Conn
	readDeadline time.Time
}

func (b *blazeHandler) SetReadDeadline(t time.Time) error {
	if err := b.conn.SetReadDeadline(t); err != nil {
		return err
	}

	b.readDeadline = t
	return nil
}

func (b *blazeHandler) ack(ctx context.Context, ackBuffer <-chan string) {
	const dur = time.Second
	t := time.NewTimer(dur)

	const maxBatch = 8 * ackBatch // 640

	requests := make([]*AcknowledgementRequest, 0, ackBatch)

	for {
		select {
		case id, ok := <-ackBuffer:
			if !ok {
				return
			}

			requests = append(requests, &AcknowledgementRequest{
				MessageID: id,
				Status:    MessageStatusRead,
			})

			if count := len(requests); count >= maxBatch {
				count = maxBatch
				if err := b.sendAcknowledgements(ctx, requests[:count]); err == nil {
					remain := requests[count:]
					copy(requests, remain)
					requests = requests[:len(remain)]

					if len(requests) == 0 {
						if !t.Stop() {
							<-t.C
						}

						t.Reset(dur)
					}
				}
			}
		case <-t.C:
			if count := len(requests); count > 0 {
				if count > maxBatch {
					count = maxBatch
				}

				if err := b.sendAcknowledgements(ctx, requests[:count]); err == nil {
					remain := requests[count:]
					copy(requests, remain)
					requests = requests[:len(remain)]
				}
			}

			t.Reset(dur)
		}
	}
}

func (b *blazeHandler) sendAcknowledgements(ctx context.Context, requests []*AcknowledgementRequest) error {
	if len(requests) <= ackBatch {
		return b.SendAcknowledgements(ctx, requests)
	}

	var g errgroup.Group
	for idx := 0; idx < len(requests); idx += ackBatch {
		right := idx + ackBatch
		if right > len(requests) {
			right = len(requests)
		}

		batch := requests[idx:right]
		g.Go(func() error {
			return b.SendAcknowledgements(ctx, batch)
		})
	}

	return g.Wait()
}
