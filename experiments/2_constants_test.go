package experiments

import (
	"fmt"
	"testing"
)

const s string = "constant"

func TestConstant(t *testing.T) {
	fmt.Println(s)

	// 常量可以定义在函数内部。
	const n string = "web3"
	fmt.Println(n)

	// 这里省略了类型，编译器会推断。
	const r = 50000
	fmt.Println(r)

	// 延伸练习：
	// 1. 增加一组常量，比较常量和变量在赋值时的差异。
	// 2. 试着引入 iota，观察连续常量的定义方式。
}
