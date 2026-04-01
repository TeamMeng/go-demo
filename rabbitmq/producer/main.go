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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 声明一个名为 "Hello" 的消息队列
	queueName := "Hello"
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	_, err = ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	ch.Close()

	// 为每个 goroutine 创建独立的 channel
	go send("Hello Big", queueName, conn)
	go send("Hello Small", queueName, conn)
	go send("Hello Midd", queueName, conn)

	time.Sleep(2 * time.Second)
}

func send(msg string, queueName string, conn *amqp.Connection) {
	// 每个 goroutine 创建独立的 channel
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to open channel: %v", err)
		return
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			Body:        []byte(msg),
			ContentType: "text/plain",
		})
}
