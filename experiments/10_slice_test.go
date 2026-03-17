package experiments

import (
	"fmt"
	"testing"
)

func TestSlice(t *testing.T) {
	var s []string
	fmt.Println(s == nil)
	fmt.Println(len(s) == 0)

	s = make([]string, 3)
	fmt.Println(s)
	fmt.Println(s != nil)
	fmt.Println(len(s))
	fmt.Println(cap(s))

	s[0] = "a"
	s[1] = "b"
	s[2] = "c"
	fmt.Println(s)

	s = append(s, "d")
	fmt.Println(s)

	a := s[0:1]
	fmt.Println(a)
}
