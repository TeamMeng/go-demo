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

	// 按字段名初始化，可读性更好。
	person := Person{Name: "Alice"}
	person.Age = 20

	fmt.Println(person)

	// 延伸练习：
	// 1. 给 Person 增加更多字段，比如 Email。
	// 2. 尝试同时写出位置初始化和字段名初始化，对比可读性。
}
