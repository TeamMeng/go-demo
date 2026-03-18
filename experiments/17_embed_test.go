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
	// 嵌入 A，表示 B 组合了 A 的字段和方法。
	A
	Address string
}

func (b B) Greet() string {
	// 这里显式复用了被嵌入类型的方法。
	return b.A.Greet() + " and I live at " + b.Address
}

func TestEmbed(t *testing.T) {
	a := A{Name: "Alice", Age: 30}
	b := B{A: a, Address: "123 Main St"}

	fmt.Println(a.Greet()) // Output: Hello, my name is Alice
	fmt.Println(b.Greet()) // Output: Hello, my name is Alice and I live at 123 Main St

	// 延伸练习：
	// 1. 直接访问 b.Name，观察嵌入字段的提升效果。
	// 2. 给 A 和 B 定义同名方法，理解方法覆盖的表现。
}
