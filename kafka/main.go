package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

func writeKafka(ctx context.Context) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  "user_click",
		Balancer:               &kafka.Hash{},
		WriteTimeout:           1 * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	for {
		for i := 0; i < 3; i++ {
			if err := writer.WriteMessages(
				ctx,
				kafka.Message{Key: []byte{1}, Value: []byte("b")},
				kafka.Message{Key: []byte{1}, Value: []byte("i")},
				kafka.Message{Key: []byte{1}, Value: []byte("g")},
			); err != nil {
				if err == kafka.LeaderNotAvailable {
					time.Sleep(500 * time.Millisecond)
					continue
				} else {
					fmt.Printf("batch write message failed: %v", err)
				}
			} else {
				break
			}
		}
		time.Sleep(time.Second)
	}
}

func main() {
	ctx := context.Background()
	go writeKafka(ctx)

	var ch chan struct{}
	ch <- struct{}{}
}
