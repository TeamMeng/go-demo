package experiments

import (
	"fmt"
	"testing"
	"time"
)

func TestBlockGoroutine(t *testing.T) {
	ch := make(chan int, 3)

	go func() {
		// 前 3 次发送会直接写入缓冲区，因为容量还没满。
		ch <- 1
		ch <- 2
		ch <- 3

		fmt.Println("ch is full")

		// 第 4 次发送时缓冲区已满。
		// 因为没有接收方把数据取走，这里会阻塞住。
		ch <- 4

		// 上面的发送已经阻塞，所以这行日志不会被执行。
		fmt.Println("this log can not be seen")
	}()

	// 留一点时间让 goroutine 跑到阻塞位置，方便直接观察输出行为。
	time.Sleep(3 * time.Second)
}
