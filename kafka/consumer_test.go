package kafka_test

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()

	consumer, err := sarama.NewConsumerGroup(addrs, "test_group", cfg)
	if err != nil {
		t.Skipf("kafka is not available: %v", err)
		return
	}
	defer consumer.Close()
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	consumer.Consume(ctx, []string{"test_topic"}, &testConsumerGroupHandler{})
	defer cancel()
}

type testConsumerGroupHandler struct {
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("Setup")
	partitions := session.Claims()["test_topic"]

	for _, part := range partitions {
		session.ResetOffset("test_topic", part, sarama.OffsetOldest, "")
	}

	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}

func (t testConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	msgs := claim.Messages()

	for msg := range msgs {
		var bizMsg MyBizMsg
		err := json.Unmarshal(msg.Value, &bizMsg)
		if err != nil {
			continue
		}
		session.MarkMessage(msg, "")
	}

	return nil
}

type MyBizMsg struct {
	Name string
}
