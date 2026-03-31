// package pubsub 演示了基于 Redis 的发布/订阅（Pub/Sub）模式的基本用法。
//
// 该包包含三个核心函数：
//   - publish: 向指定 Redis 频道发布消息。
//   - subscribe: 订阅一个或多个 Redis 频道并阻塞接收消息。
//   - pubSub: 编排演示流程，先启动订阅者，再启动发布者。
//
// 注意：
//
//	本包中的函数均通过 context.Context 传递发布者/订阅者的名称，仅用于日志打印。
//	实际生产环境中建议使用结构化的日志库（如 slog、zap）替代 fmt.Printf。
package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// publish 向指定的 Redis 频道发布一条消息。
//
// 参数：
//   - ctx: 上下文对象，预期包含 "publisher_name" 键值对，用于日志标识。
//   - client: 已连接的 Redis 客户端实例。
//   - channel: 目标频道名称。
//   - message: 要发布的消息内容，支持任意类型（底层由 go-redis 序列化）。
//
// 行为说明：
//
//	调用 client.Publish 执行发布操作，并根据返回值打印成功或失败日志。
//	成功时会输出当前时刻订阅该频道的在线客户端数量。
func publish(ctx context.Context, client *redis.Client, channel string, message any) {
	cmd := client.Publish(ctx, channel, message)
	if cmd.Err() == nil {
		n := cmd.Val()
		fmt.Printf("%s向频道%s发布了信息,此时该频道有%d个订阅者\n", ctx.Value("publisher_name"), channel, n)
	} else {
		fmt.Printf("%s向频道%s发布信息失败%v\n", ctx.Value("publisher_name"), channel, cmd.Err())
	}
}

// subscribe 订阅一个或多个 Redis 频道，并进入阻塞循环接收消息。
//
// 参数：
//   - ctx: 上下文对象，预期包含 "subscriber_name" 键值对，用于日志标识。
//   - client: 已连接的 Redis 客户端实例。
//   - channels: 要订阅的频道名称列表，支持同时订阅多个频道。
//
// 行为说明：
//  1. 调用 client.Subscribe 创建 PubSub 订阅器，并在函数返回时通过 defer 关闭它。
//  2. 进入无限循环，调用 ps.ReceiveMessage 阻塞等待新消息。
//  3. 收到消息后打印订阅者名称、频道名称和消息内容。
//  4. 若接收过程中发生错误（如上下文取消、连接断开），打印错误并退出循环。
//
// 注意：
//
//	该函数应在独立的 goroutine 中运行，否则会阻塞调用方。
func subscribe(ctx context.Context, client *redis.Client, channels []string) {
	ps := client.Subscribe(ctx, channels...)
	defer ps.Close()

	for {
		if msg, err := ps.ReceiveMessage(ctx); err != nil {
			fmt.Println(err)
			break
		} else {
			fmt.Printf("%s从频道%s里收到信息%s\n", ctx.Value("subscriber_name"), msg.Channel, msg.Payload)
		}
	}
}

// pubSub 编排整个 Redis Pub/Sub 演示流程。
//
// 执行流程：
//  1. 构造两个发布者上下文（publisher1、publisher2）和两个订阅者上下文（subscriber1、subscriber2）。
//  2. 定义两个频道 channel1 和 channel2。
//  3. 在独立 goroutine 中启动两个订阅者：
//     - subscriber1 订阅 channel1
//     - subscriber2 订阅 channel2
//     启动后休眠 1 秒，确保订阅者已连接到 Redis 并注册频道。
//  4. 在独立 goroutine 中启动两个发布者：
//     - publisher1 向 channel1 发布 "白日依山尽"
//     - publisher2 向 channel2 发布 "黄河入海流"
//     启动后休眠 1 秒，给消息传输和日志输出留出时间。
//
// 注意：
//
//	该函数仅用于演示，休眠时间（time.Sleep）在实际生产代码中应避免使用。
//	若需要精确控制生命周期，建议使用 sync.WaitGroup 或 channel 进行 goroutine 同步。
func pubSub(ctx context.Context, client *redis.Client) {
	ctx1 := context.WithValue(ctx, "publisher_name", "publisher1")
	ctx2 := context.WithValue(ctx, "publisher_name", "publisher2")

	channel1 := "channel1"
	channel2 := "channel2"

	ctx3 := context.WithValue(ctx, "subscriber_name", "subscriber1")
	ctx4 := context.WithValue(ctx, "subscriber_name", "subscriber2")

	// 启动订阅者 goroutine，分别监听不同频道
	go subscribe(ctx3, client, []string{channel1})
	go subscribe(ctx4, client, []string{channel2})
	time.Sleep(1 * time.Second)

	// 启动发布者 goroutine，向对应频道发送消息
	go publish(ctx1, client, channel1, "白日依山尽")
	go publish(ctx2, client, channel2, "黄河入海流")
	time.Sleep(1 * time.Second)

	ctx5 := context.WithValue(ctx, "subscriber_name", "subscriber3")
	go subscribe(ctx5, client, []string{channel1, channel2})
	time.Sleep(1 * time.Second)

	go publish(ctx1, client, channel2, "123456")
	go publish(ctx2, client, channel2, "7891011")
	time.Sleep(1 * time.Second)

}
