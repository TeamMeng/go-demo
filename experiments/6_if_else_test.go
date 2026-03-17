package experiments

import (
	"fmt"
	"testing"
)

func TestIfElse(t *testing.T) {
	a := 10
	if a > 5 {
		fmt.Println("a is greater than 5")
	} else {
		fmt.Println("a is not greater than 5")
	}
}
