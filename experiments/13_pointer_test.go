package experiments

import (
	"fmt"
	"testing"
)

func Pointer(ptr *int) {
	*ptr = 2
}

func Value(val int) {
	val = 2
}

func TestPointer(t *testing.T) {
	i := 1
	fmt.Println(i)
	Pointer(&i)
	fmt.Println(i)

	i = 1
	fmt.Println(i)
	Value(i)
	fmt.Println(i)
}
