package experiments

import (
	"fmt"
	"testing"
)

func TestType(t *testing.T) {
	// 1. Struct
	type Person struct {
		Name string
		Age  int
	}

	var person Person

	fmt.Println(person)

	// 2. Interface
	type Animal interface {
		gener() string
		color() string
	}
}
