package experiments

import (
	"fmt"
	"testing"
	"time"
)

func printlnHelloWorld() {
	fmt.Println("Hello, World!")
}

func TestGoroutine(t *testing.T) {
	// 使用 go 关键字启动并发任务。
	go printlnHelloWorld()

	// 匿名函数也可以直接作为 goroutine 启动。
	go func() {
		fmt.Println("Hello again!")
	}()

	// 教学示例里用 Sleep 等待 goroutine 执行完成。
	time.Sleep(time.Second)

	// 延伸练习：
	// 1. 把 Sleep 替换成 WaitGroup。
	// 2. 同时启动多个 goroutine，观察输出顺序是否固定。
}
