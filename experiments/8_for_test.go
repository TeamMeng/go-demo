package experiments

import (
	"fmt"
	"testing"
)

func TestFor(t *testing.T) {
	// 传统 for 循环。
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}

	i := 1
	// 类似 while 的写法。
	for i < 10 {
		fmt.Println(i)
		i += 1
	}

	// range 一个整数会产生 0 到 n-1。
	for i := range 3 {
		fmt.Println(i)
	}

	// 无限循环，通常配合 break 退出。
	for {
		fmt.Println("Infinite loop")
		break
	}

	// 延伸练习：
	// 1. 用 for 遍历一个 slice。
	// 2. 把无限循环改成带条件退出的读取模型。
}
