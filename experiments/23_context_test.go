package experiments

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	// WithCancel 返回一个可取消的上下文和对应的 cancel 函数。
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		// Done 在上下文被取消时关闭。
		case <-ctx.Done():
			fmt.Println("context canceled")
		case <-time.After(2 * time.Second):
			fmt.Println("timeout")
		}
	}()

	time.Sleep(time.Second)

	// ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	// go func() {
	// 	for range 50 {
	// 		select {
	// 		case <-ctx.Done():
	// 			fmt.Println("context canceled")
	// 		case <-time.After(2 * time.Second):
	// 			fmt.Println("timeout")
	// 		}
	// 	}

	// }()

	// time.Sleep(time.Second * 3)

	// 延伸练习：
	// 1. 把 WithCancel 改成 WithTimeout，观察自动超时。
	// 2. 在多个 goroutine 间共享同一个 ctx，统一取消它们。
}
