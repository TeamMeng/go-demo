package experiments

import (
	"fmt"
	"testing"
)

func TestChannel(t *testing.T) {
	// 无缓冲 channel：发送和接收需要配对同步。
	ch := make(chan int)

	go func() {
		val := <-ch
		fmt.Println(val)
	}()

	ch <- 10

	// 有缓冲 channel：在容量未满前，发送可以先完成。
	ch = make(chan int, 10)
	ch <- 10

	// 延伸练习：
	// 1. 再从缓冲 channel 读取一次，观察先进先出行为。
	// 2. 尝试关闭 channel，并思考关闭后的读取结果。
}
