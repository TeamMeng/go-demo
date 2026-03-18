package experiments

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
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

}
