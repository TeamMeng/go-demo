package experiments

import (
	"fmt"
	"testing"
)

func Pointer(ptr *int) {
	// 通过指针可以修改调用方持有的原始值。
	*ptr = 2
}

func Value(val int) {
	// 这里只修改了副本，不会影响外部变量。
	val = 2
}

func TestPointer(t *testing.T) {
	i := 1
	fmt.Println(i)
	Pointer(&i)
	fmt.Println(i)

	i = 1
	fmt.Println(i)
	Value(i)
	fmt.Println(i)

	// 延伸练习：
	// 1. 把 int 换成 struct，继续观察传值和传指针的区别。
	// 2. 尝试返回指针，并思考它适合什么场景。
}
