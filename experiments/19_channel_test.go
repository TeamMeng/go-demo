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

	ch = make(chan int, 2)
	ch <- 123
	val, ok := <-ch
	fmt.Printf("val: %v, ok: %v\n", val, ok)
	ch <- 234
	close(ch)
	val, ok = <-ch
	fmt.Printf("val: %v, ok: %v\n", val, ok)
}
