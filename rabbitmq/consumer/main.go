package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// main 函数是 RabbitMQ 消息消费者的入口
// 该程序用于连接 RabbitMQ 服务器，声明队列，并持续监听接收消息
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
	// 参数说明：name="Hello"(队列名), durable=false(不持久化), autoDelete=false(不自动删除),
	// exclusive=false(非独占), noWait=false(等待服务器响应), args=nil(额外参数)
	// 注意：如果队列已存在，此操作不会有任何影响；如果不存在，则会创建新队列
	ch.QueueDeclare(queueName, false, false, false, false, nil)

	go receive(queueName, ch, 1)
	go receive(queueName, ch, 2)
	go receive(queueName, ch, 3)
	var block chan struct{}
	block <- struct{}{}
}

func receive(queueName string, ch *amqp.Channel, flag int) {

	// 开始消费队列中的消息
	// 参数说明：queue="Hello"(队列名), consumer=""(消费者标签，自动生成),
	// autoAck=false(不自动确认), exclusive=false(非独占), noLocal=false(接收本地消息),
	// noWait=false(等待服务器响应), args=nil(额外参数)
	deliveryCh, _ := ch.Consume(queueName, "", false, false, false, false, nil)

	// 使用 for-range 循环持续监听消息通道
	// 当有消息到达时，会打印消息内容到日志
	for delivery := range deliveryCh {
		log.Printf("[%d] receive message [%s]", flag, delivery.Body)
		delivery.Ack(false)
	}
}
