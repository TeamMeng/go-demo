package experiments

import (
	"fmt"
	"testing"
)

func TestVariable(t *testing.T) {
	// var 方式声明，类型由右值推断。
	var a = "initial"
	fmt.Println(a)

	// := 是函数内部最常见的短变量声明。
	f := "apple"
	fmt.Println(f)

	// 延伸练习：
	// 1. 再声明一个 int 和 bool 变量，观察零值。
	// 2. 对比 var 和 := 在函数外/函数内的使用限制。
}
