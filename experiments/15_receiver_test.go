package experiments

import (
	"fmt"
	"testing"
)

type person struct {
	name string
	age  int
}

func (p person) valueUpdate() {
	p.age += 1
}

func (p *person) pointerUpdate() {
	p.age += 1
}

func TestReceiver(t *testing.T) {
	p := person{name: "Alice", age: 30}
	p.valueUpdate()
	fmt.Println(p)

	p.pointerUpdate()
	fmt.Println(p)
}
