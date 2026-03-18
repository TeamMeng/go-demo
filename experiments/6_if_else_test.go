package experiments

import (
	"fmt"
	"testing"
)

func TestIfElse(t *testing.T) {
	a := 10
	// Go 的条件表达式不需要括号。
	if a > 5 {
		fmt.Println("a is greater than 5")
	} else {
		fmt.Println("a is not greater than 5")
	}

	// 延伸练习：
	// 1. 增加 else if 分支，细分多个区间。
	// 2. 试试 if 前置语句的写法。
}
