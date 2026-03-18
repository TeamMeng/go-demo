package experiments

import (
	"fmt"
	"testing"
	"time"
)

func printlnHelloWord() {
	fmt.Println("Hello, World!")
}

func TestGoRoutine(t *testing.T) {
	go printlnHelloWord()

	go func() {
		fmt.Println("Hello again!")
	}()

	time.Sleep(time.Second)
}
