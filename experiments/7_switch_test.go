package experiments

import (
	"fmt"
	"testing"
	"time"
)

func TestSwitch(t *testing.T) {
	i := 2
	// 普通值匹配 switch。
	switch i {
	case 1:
		fmt.Println("i is 1")
	case 2:
		fmt.Println("i is 2")
	default:
		fmt.Println("i is something else")
	}

	// 无表达式 switch，常用于替代一长串 if / else if。
	now := time.Now()
	switch {
	case now.Hour() < 12:
		fmt.Println("Good morning!")
	case now.Hour() < 18:
		fmt.Println("Good afternoon!")
	default:
		fmt.Println("Good evening!")
	}

	// 延伸练习：
	// 1. 给普通 switch 增加多个 case 合并匹配。
	// 2. 试试 switch 中使用 fallthrough 的效果。
}
