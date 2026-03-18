package experiments

import (
	"fmt"
	"sync"
	"testing"
)

type Container struct {
	mu    sync.Mutex
	count map[string]int
}

func (c *Container) inc(name string) {
	// 对共享 map 的写操作需要加锁保护。
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count[name]++
}

func TestMutex(t *testing.T) {

	c := Container{
		count: map[string]int{"a": 0, "b": 0},
	}
	var wg sync.WaitGroup

	var update = func(name string, n int) {
		// range 一个整数会执行 n 次。
		for range n {
			c.inc(name)
		}
	}

	// 新版 WaitGroup 提供了 Go 方法，便于直接启动任务。
	wg.Go(func() { update("a", 100) })
	wg.Go(func() { update("b", 200) })

	wg.Wait()

	fmt.Println(c.count["a"], c.count["b"])

	// 延伸练习：
	// 1. 同时启动多个 goroutine 更新同一个 key。
	// 2. 去掉锁后用 race detector 观察会发生什么。
}
