package experiments

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrencyWithoutLock(t *testing.T) {
	var shared []int
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			shared = append(shared, 1)
		}()
	}

	wg.Wait()

	fmt.Println(len(shared))
}

func TestConcurrencyWithLock(t *testing.T) {
	var shared []int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			shared = append(shared, 1)
			mu.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println(len(shared))
}
