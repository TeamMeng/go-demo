package experiments

import (
	"fmt"
	"testing"
)

func TestFor(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}

	i := 1
	for i < 10 {
		fmt.Println(i)
		i += 1
	}

	for i := range 3 {
		fmt.Println(i)
	}

	for {
		fmt.Println("Infinite loop")
		break
	}
}
