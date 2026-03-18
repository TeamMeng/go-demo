package experiments

import (
	"fmt"
	"testing"
)

func TestType(t *testing.T) {
	// Struct 用于组织一组字段。
	type Person struct {
		Name string
		Age  int
	}

	var person Person

	fmt.Println(person)

	// Interface 用于描述一组行为。
	type Animal interface {
		kind() string
		color() string
	}

	// 延伸练习：
	// 1. 定义一个 Dog 结构体，实现 Animal 接口。
	// 2. 新增一个函数，接收 Animal 并打印它的行为。
}
