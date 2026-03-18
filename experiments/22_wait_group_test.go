package experiments

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func worker(id int) {
	fmt.Printf("Worker %d starting\n", id)

	time.Sleep(time.Second)
	fmt.Printf("Worker %d done\n", id)
}

func TestWaitGroup(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		id := i
		// WaitGroup 负责等待一组 goroutine 全部结束。
		wg.Go(func() {
			worker(id)
		})
	}
	wg.Wait()

	// 延伸练习：
	// 1. 把循环变量作为参数传进闭包，比较输出差异。
	// 2. 让每个 worker 返回结果，再思考 WaitGroup 不负责结果收集这一点。
}
