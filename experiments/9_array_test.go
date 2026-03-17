package experiments

import (
	"fmt"
	"testing"
)

func TestArray(t *testing.T) {
	var a [5]int
	fmt.Println(a)

	a[4] = 100
	fmt.Println(a)
	fmt.Println("len: ", len(a))

	b := [5]int{1, 2, 3, 4, 5}
	fmt.Println(b)

	b = [...]int{1, 2, 3, 4, 5}
	fmt.Println(b)

	c := [...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(c)
	fmt.Println("len: ", len(c))
}
