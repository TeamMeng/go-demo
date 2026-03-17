package experiments

import (
	"fmt"
	"testing"
)

func TestStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	person := Person{Name: "Alice"}
	person.Age = 20

	fmt.Println(person)
}
