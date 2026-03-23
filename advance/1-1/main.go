package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type ChatRoom struct {
	// clients 保存当前在线连接及其用户名。
	clients map[net.Conn]string
	// mutex 保护 clients，避免多个 goroutine 并发读写 map。
	mutex sync.Mutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients: make(map[net.Conn]string),
	}
}

func (c *ChatRoom) broadcast(sender net.Conn, message string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for client := range c.clients {
		// 广播时跳过发送者自己，只把消息转发给其他在线用户。
		if sender == client {
			continue
		}
		client.Write([]byte(message))
	}
}

func (c *ChatRoom) handleConnection(conn net.Conn) {
	// 每个客户端连接都由一个独立 goroutine 处理，结束时关闭连接。
	defer conn.Close()

	conn.Write([]byte("Please input your name: "))
	name := make([]byte, 1024)
	n, _ := conn.Read(name)
	userName := strings.TrimSpace(string(name[:n]))

	c.mutex.Lock()
	// 完成握手后把用户注册到聊天室。
	c.clients[conn] = userName
	c.mutex.Unlock()

	c.broadcast(conn, fmt.Sprintf("system: welcome %v joined the chat room\n", userName))
	for {
		readBuf := make([]byte, 1024)
		n, err := conn.Read(readBuf)
		if err != nil {
			break
		}
		message := strings.TrimSpace(string(readBuf[:n]))
		if message == "quit" {
			break
		}
		// 普通消息按“用户名: 内容”的格式转发给其他客户端。
		c.broadcast(conn, fmt.Sprintf("%v: %s\n", userName, message))
	}

	c.mutex.Lock()
	delete(c.clients, conn)
	c.mutex.Unlock()

	c.broadcast(conn, fmt.Sprintf("system: %v quit the chat room\n", userName))
}

func main() {
	chatRoom := NewChatRoom()
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("Server listen failed")
		return
	}
	defer listener.Close()
	fmt.Println("Server listening...")

	for {
		conn, _ := listener.Accept()
		go chatRoom.handleConnection(conn)
	}
}
