package experiments

import (
	"fmt"
	"testing"
)

func TestChan(t *testing.T) {
	ch := make(chan int)

	go func() {
		val := <-ch
		fmt.Println(val)
	}()

	ch <- 10

	ch = make(chan int, 10)
	ch <- 10
}
