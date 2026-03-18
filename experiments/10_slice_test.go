package experiments

import (
	"fmt"
	"testing"
)

func TestSlice(t *testing.T) {
	// 零值切片是 nil，但依然可以安全地拿来读 len。
	var s []string
	fmt.Println(s == nil)
	fmt.Println(len(s) == 0)

	// make 为切片分配底层数组。
	s = make([]string, 3)
	fmt.Println(s)
	fmt.Println(s != nil)
	fmt.Println(len(s))
	fmt.Println(cap(s))

	s[0] = "a"
	s[1] = "b"
	s[2] = "c"
	fmt.Println(s)

	// append 可能复用原数组，也可能触发扩容。
	s = append(s, "d")
	fmt.Println(s)

	// 切片表达式返回的是一个视图，不是深拷贝。
	a := s[0:1]
	fmt.Println(a)

	// 延伸练习：
	// 1. 打印 append 前后的 cap，观察扩容变化。
	// 2. 修改子切片中的元素，观察原切片是否变化。
}
