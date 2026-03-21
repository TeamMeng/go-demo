package experiments

import (
	"fmt"
	"sync"
	"testing"
)

// TestConcurrencyWithoutLock 演示多个 goroutine 并发写入同一个 slice 时，
// 如果没有任何同步保护，会产生数据竞争，长度结果也可能不稳定。
func TestConcurrencyWithoutLock(t *testing.T) {
	if isRaceEnabled {
		t.Skip("skip intentional data race demo when running with -race")
	}

	var shared []int
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// append 会修改 slice 的底层结构和长度，并发执行时并不是线程安全的。
			shared = append(shared, 1)
		}()
	}

	wg.Wait()

	// 输出结果仅用于观察；若配合 `go test -race`，这里通常可以看到数据竞争。
	fmt.Println(len(shared))
}

// TestConcurrencyWithLock 演示对共享 slice 的写入加锁后，
// 可以把临界区串行化，从而避免并发写入导致的数据竞争。
func TestConcurrencyWithLock(t *testing.T) {
	var shared []int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 进入临界区前先加锁，确保同一时刻只有一个 goroutine 修改 shared。
			mu.Lock()
			shared = append(shared, 1)
			mu.Unlock()
		}()
	}

	wg.Wait()

	// 在加锁保护下，理论上这里应稳定输出 10。
	fmt.Println(len(shared))
}
