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
	// 值接收者拿到的是副本，外部对象不会变。
	p.age += 1
}

func (p *person) pointerUpdate() {
	// 指针接收者可以直接修改原对象。
	p.age += 1
}

func TestReceiver(t *testing.T) {
	p := person{name: "Alice", age: 30}
	p.valueUpdate()
	fmt.Println(p)

	p.pointerUpdate()
	fmt.Println(p)

	// 延伸练习：
	// 1. 给 person 增加一个只读方法，比如 String。
	// 2. 观察当结构体变大时，值接收者的拷贝成本意味着什么。
}
