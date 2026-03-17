package experiments

import (
	"fmt"
	"testing"
)

func TestMap(t *testing.T) {
	m := make(map[string]int)
	fmt.Println(m)

	m["one"] = 1
	m["two"] = 2
	fmt.Println(m)

	value := m["one"]
	fmt.Println(value)

	delete(m, "two")
	fmt.Println(m)
	value = m["two"]
	fmt.Println(value)

	clear(m)
	fmt.Println(m)
}
