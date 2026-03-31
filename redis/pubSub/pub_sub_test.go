// package pubsub 包含 Redis Pub/Sub 功能的单元测试。
//
// 测试使用 miniredis 作为内存中的 Redis 服务器，无需外部 Redis 实例即可运行。
// 所有测试都是独立的，可以并行执行（通过 t.Parallel()）。
package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis 创建一个内存中的 Redis 服务器并返回客户端和清理函数。
//
// 每个测试应该独立调用此函数，确保测试之间的隔离性。
func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	// 创建 miniredis 服务器
	s := miniredis.RunT(t)

	// 创建 redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// 返回清理函数，用于关闭客户端和服务器
	cleanup := func() {
		client.Close()
		s.Close()
	}

	return client, cleanup
}

// TestPublish 测试发布消息功能。
//
// 验证点：
//   - 向频道发布消息时，返回的订阅者数量正确（没有订阅者时应返回 0）。
//   - 发布操作本身不会返回错误。
func TestPublish(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.WithValue(context.Background(), "publisher_name", "test_publisher")
	channel := "test_channel"
	message := "hello world"

	// 测试：向没有订阅者的频道发布消息
	cmd := client.Publish(ctx, channel, message)
	if cmd.Err() != nil {
		t.Fatalf("publish failed: %v", cmd.Err())
	}

	// 没有订阅者，应该返回 0
	if cmd.Val() != 0 {
		t.Errorf("expected 0 subscribers, got %d", cmd.Val())
	}
}

// TestSubscribe 测试订阅和接收消息功能。
//
// 验证点：
//   - 订阅者能够成功订阅频道。
//   - 订阅者能够接收到发布到该频道的消息。
//   - 消息内容、频道名称正确。
func TestSubscribe(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channel := "test_channel"
	expectedMessage := "test_message"

	// 用于同步测试的 channel
	receivedMsg := make(chan string, 1)
	done := make(chan struct{})

	// 在 goroutine 中启动订阅者
	go func() {
		defer close(done)

		ps := client.Subscribe(ctx, channel)
		defer ps.Close()

		// 等待消息
		msg, err := ps.ReceiveMessage(ctx)
		if err != nil {
			t.Errorf("receive message failed: %v", err)
			return
		}

		receivedMsg <- msg.Payload
	}()

	// 等待订阅者建立连接
	time.Sleep(100 * time.Millisecond)

	// 发布消息
	client.Publish(ctx, channel, expectedMessage)

	// 等待接收消息或超时
	select {
	case payload := <-receivedMsg:
		if payload != expectedMessage {
			t.Errorf("expected message %q, got %q", expectedMessage, payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	cancel() // 取消上下文，让订阅者退出
	<-done   // 等待 goroutine 结束
}

// TestPubSubIntegration 测试完整的发布/订阅集成流程。
//
// 验证点：
//   - 多个订阅者可以订阅不同频道。
//   - 发布者向特定频道发送的消息只被该频道的订阅者接收。
//   - 消息内容在传输过程中保持一致。
func TestPubSubIntegration(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channel1 := "channel1"
	channel2 := "channel2"
	msg1 := "message for channel1"
	msg2 := "message for channel2"

	// 用于接收消息的 channel
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	done := make(chan struct{}, 2)

	// 订阅者 1：监听 channel1
	go func() {
		defer func() { done <- struct{}{} }()

		ps := client.Subscribe(ctx, channel1)
		defer ps.Close()

		msg, err := ps.ReceiveMessage(ctx)
		if err != nil {
			return // 上下文取消是预期的行为
		}
		ch1 <- msg.Payload
	}()

	// 订阅者 2：监听 channel2
	go func() {
		defer func() { done <- struct{}{} }()

		ps := client.Subscribe(ctx, channel2)
		defer ps.Close()

		msg, err := ps.ReceiveMessage(ctx)
		if err != nil {
			return // 上下文取消是预期的行为
		}
		ch2 <- msg.Payload
	}()

	// 等待订阅者建立连接
	time.Sleep(100 * time.Millisecond)

	// 发布消息
	client.Publish(ctx, channel1, msg1)
	client.Publish(ctx, channel2, msg2)

	// 验证 channel1 收到正确的消息
	select {
	case payload := <-ch1:
		if payload != msg1 {
			t.Errorf("channel1: expected %q, got %q", msg1, payload)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for channel1 message")
	}

	// 验证 channel2 收到正确的消息
	select {
	case payload := <-ch2:
		if payload != msg2 {
			t.Errorf("channel2: expected %q, got %q", msg2, payload)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for channel2 message")
	}

	// 清理
	cancel()
	<-done
	<-done
}

// TestMultipleSubscribersSameChannel 测试同一频道的多个订阅者。
//
// 验证点：
//   - 同一频道的多个订阅者都能收到消息（Pub/Sub 是广播模式）。
//   - 发布者返回的订阅者数量正确。
func TestMultipleSubscribersSameChannel(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channel := "shared_channel"
	message := "broadcast message"

	receivedCount := make(chan int, 1)
	done := make(chan struct{}, 2)

	// 启动两个订阅者
	for i := 0; i < 2; i++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()

			ps := client.Subscribe(ctx, channel)
			defer ps.Close()

			msg, err := ps.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			if msg.Payload == message {
				receivedCount <- 1
			}
		}(i)
	}

	// 等待订阅者建立连接
	time.Sleep(100 * time.Millisecond)

	// 发布消息，应该有 2 个订阅者
	cmd := client.Publish(ctx, channel, message)
	if cmd.Err() != nil {
		t.Fatalf("publish failed: %v", cmd.Err())
	}

	// 验证订阅者数量
	if cmd.Val() != 2 {
		t.Errorf("expected 2 subscribers, got %d", cmd.Val())
	}

	// 统计实际收到的消息数量
	count := 0
	timeout := time.After(2 * time.Second)
loop:
	for {
		select {
		case <-receivedCount:
			count++
			if count == 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if count != 2 {
		t.Errorf("expected 2 messages received, got %d", count)
	}

	cancel()
	<-done
	<-done
}

// TestSubscribeMultipleChannels 测试单个订阅者订阅多个频道。
//
// 验证点：
//   - 一个订阅者可以同时监听多个频道。
//   - 订阅者能正确接收来自不同频道的消息。
func TestSubscribeMultipleChannels(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channels := []string{"ch1", "ch2", "ch3"}
	received := make(map[string]bool)
	mu := make(chan struct{}, 1) // 用于保护 received map 的简单互斥

	done := make(chan struct{})

	// 启动订阅者监听多个频道
	go func() {
		defer close(done)

		ps := client.Subscribe(ctx, channels...)
		defer ps.Close()

		for i := 0; i < len(channels); i++ {
			msg, err := ps.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			mu <- struct{}{}
			received[msg.Channel] = true
			<-mu
		}
	}()

	// 等待订阅者建立连接
	time.Sleep(100 * time.Millisecond)

	// 向每个频道发送消息
	for _, ch := range channels {
		client.Publish(ctx, ch, "msg")
	}

	// 等待接收或超时
	select {
	case <-done:
		// 成功接收所有消息
	case <-time.After(2 * time.Second):
		// 检查实际收到了多少
		mu <- struct{}{}
		count := len(received)
		<-mu
		if count != len(channels) {
			t.Errorf("expected %d channels received, got %d", len(channels), count)
		}
	}

	cancel()
}
