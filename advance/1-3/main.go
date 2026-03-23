package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 示例中允许所有来源建立连接，真实项目里应按域名白名单校验。
		return true
	},
}

func main() {
	// 注册 WebSocket 入口，客户端通过 ws://host:8080/ws 建立长连接。
	http.HandleFunc("/ws", handleWebSocket)

	http.ListenAndServe(":8080", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 先将普通 HTTP 请求升级成 WebSocket 双向连接。
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}
	// 处理结束后关闭连接，释放底层网络资源。
	defer conn.Close()

	fmt.Println("New WebSocket Conn")
	for {
		// 持续读取客户端发来的消息，直到连接断开或读取失败。
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		fmt.Printf("Received message: %s\n", message)
		// 在原消息末尾追加确认文本，再按原消息类型回写给客户端。
		message = append(message, "已收到"...)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
	fmt.Println("WebSocket 连接已关闭")
}
