package experiments

import (
	"fmt"
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(time.Second * 2)
		ch1 <- "message1"
	}()

	go func() {
		time.Sleep(time.Second * 1)
		ch2 <- "message2"
	}()

	counter := 0

	for {
		select {
		case message1 := <-ch1:
			fmt.Println(message1)
			counter++
		case message2 := <-ch2:
			fmt.Println(message2)
			counter++
		default:
			if counter == 2 {
				return
			}
		}
	}
}
