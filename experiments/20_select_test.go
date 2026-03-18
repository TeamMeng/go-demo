package experiments

import (
	"fmt"
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(time.Second * 2)
		ch1 <- "message1"
	}()

	go func() {
		time.Sleep(time.Second * 1)
		ch2 <- "message2"
	}()

	counter := 0

	for {
		// select 会在多个可通信的 case 中选择一个执行。
		select {
		case message1 := <-ch1:
			fmt.Println(message1)
			counter++
		case message2 := <-ch2:
			fmt.Println(message2)
			counter++
		default:
			// default 让 select 变成非阻塞，这里会持续轮询直到两个消息都收到。
			if counter == 2 {
				return
			}
		}
	}

	// 延伸练习：
	// 1. 去掉 default，观察阻塞式 select 的行为。
	// 2. 加入 time.After 分支，实现超时退出。
}
