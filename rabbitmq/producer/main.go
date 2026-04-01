package main

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// main 函数是 RabbitMQ 消息生产者的入口
// 该程序用于连接 RabbitMQ 服务器，声明一个队列，并向该队列发送消息
func main() {
	// 连接到 RabbitMQ 服务器，使用默认的 guest/guest 凭据
	// 连接地址为本地主机的 5672 端口（RabbitMQ 默认端口）
	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672")
	defer conn.Close()

	// 在连接上创建一个通信通道（Channel）
	// Channel 是轻量级的连接，用于实际的 AMQP 操作
	ch, _ := conn.Channel()
	defer ch.Close()

	// 声明一个名为 "Hello" 的消息队列
	queueName := "Hello"
	// 参数说明：name="Hello"(队列名), durable=true(持久化), autoDelete=true(自动删除),
	// exclusive=false(非独占), noWait=false(等待服务器响应), args=nil(额外参数)
	_, err := ch.QueueDeclare(queueName, true, true, false, false, nil)
	if err != nil {
		log.Panic(err)
	}

	go send("Hello Big", queueName, ch)
	go send("Hello Small", queueName, ch)
	go send("Hello Midd", queueName, ch)

	time.Sleep(2 * time.Second)
}

func send(msg string, queueName string, ch *amqp.Channel) {
	// 在连接上创建一个通信通道（Channel）
	// Channel 是轻量级的连接，用于实际的 AMQP 操作
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 发布消息到指定的队列
	// 参数说明：ctx(上下文), exchange=""(使用默认交换机), routingKey="Hello"(路由键/队列名),
	// mandatory=false(不强制的), immediate=false(不立即的), Publishing(消息内容)
	ch.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			Body:        []byte(msg),  // 消息体，转换为字节数组
			ContentType: "text/plain", // 内容类型为纯文本
		})
}
