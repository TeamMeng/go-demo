package main

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// main 函数是 RabbitMQ 消息消费者的入口
// 该程序用于连接 RabbitMQ 服务器，声明队列，并持续监听接收消息
func main() {
	// 连接到 RabbitMQ 服务器，使用默认的 guest/guest 凭据
	// 连接地址为本地主机的 5672 端口（RabbitMQ 默认端口）
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 声明一个名为 "Hello" 的消息队列
	queueName := "Hello"
	ch, _ := conn.Channel()
	ch.QueueDeclare(queueName, false, false, false, false, nil)
	ch.Close()

	// 为每个消费者创建独立的 channel（AMQP channel 不是线程安全的）
	for i := 1; i <= 3; i++ {
		go receive(queueName, conn, i)
	}

	// 阻塞主进程，防止退出
	select {}
}

func receive(queueName string, conn *amqp.Connection, flag int) {
	// 每个消费者创建独立的 channel
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[%d] failed to open channel: %v", flag, err)
		return
	}
	defer ch.Close()

	// 设置 prefetch count，每个消费者最多缓存 2 条未确认的消息
	// 这样能实现真正的负载均衡：处理快的消费者会优先收到更多消息
	err = ch.Qos(2, 0, false)
	if err != nil {
		log.Printf("[%d] failed to set QoS: %v", flag, err)
		return
	}

	// 开始消费队列中的消息，autoAck=false 改为手动确认
	deliveryCh, err := ch.Consume(
		queueName,
		"",
		false, // autoAck: false 表示手动确认，确保消息被处理后才从队列移除
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		log.Printf("[%d] failed to register consumer: %v", flag, err)
		return
	}

	// 使用 for-range 循环持续监听消息通道
	for delivery := range deliveryCh {
		log.Printf("[%d] receive message [%s]", flag, delivery.Body)
		// 模拟处理耗时
		time.Sleep(100 * time.Millisecond)
		// 手动确认消息已处理
		delivery.Ack(false)
	}
}
