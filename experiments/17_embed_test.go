package experiments

import (
	"fmt"
	"testing"
)

type A struct {
	Name string
	Age  int
}

func (a A) Greet() string {
	return "Hello, my name is " + a.Name
}

type B struct {
	A
	Address string
}

func (b B) Greet() string {
	return b.A.Greet() + " and I live at " + b.Address
}

func TestEmbed(t *testing.T) {
	a := A{Name: "Alice", Age: 30}
	b := B{A: a, Address: "123 Main St"}

	fmt.Println(a.Greet()) // Output: Hello, my name is Alice
	fmt.Println(b.Greet()) // Output: Hello, my name is Alice and I live at 123 Main St
}
