package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/bytedance/gopkg/util/logger"
)

type BatchHandler[T any] struct {
	l             logger.Logger
	fn            func(msg []*sarama.ConsumerMessage, ts []T) error
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](
	l logger.Logger,
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error,
) *BatchHandler[T] {
	return &BatchHandler[T]{l: l, fn: fn, batchDuration: time.Second, batchSize: 10}
}

// Setup implements [sarama.ConsumerGroupHandler].
func (b *BatchHandler[T]) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements [sarama.ConsumerGroupHandler].
func (b *BatchHandler[T]) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim implements [sarama.ConsumerGroupHandler].
func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
		var last *sarama.ConsumerMessage
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, b.batchSize)
		ts := make([]T, 0, b.batchSize)
		for i := 0; i < b.batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					cancel()
					return nil
				}
				last = msg
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("serde failed")
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		err := b.fn(msgs, ts)
		if err != nil {
			b.l.Error(err)
			continue
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}
