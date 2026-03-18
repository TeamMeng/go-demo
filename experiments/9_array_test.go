package experiments

import (
	"fmt"
	"testing"
)

func TestArray(t *testing.T) {
	// 数组长度是类型的一部分，这里是 [5]int。
	var a [5]int
	fmt.Println(a)

	a[4] = 100
	fmt.Println(a)
	fmt.Println("len: ", len(a))

	b := [5]int{1, 2, 3, 4, 5}
	fmt.Println(b)

	b = [...]int{1, 2, 3, 4, 5}
	fmt.Println(b)

	// ... 让编译器根据元素数量推断数组长度。
	c := [...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(c)
	fmt.Println("len: ", len(c))

	// 延伸练习：
	// 1. 定义二维数组并打印。
	// 2. 对比数组传参和切片传参的行为差异。
}
