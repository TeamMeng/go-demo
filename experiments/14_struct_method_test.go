package experiments

import (
	"fmt"
	"testing"
)

type rect struct {
	width, height int
}

func (r *rect) area() int {
	// 指针接收者常用于避免拷贝或允许后续修改状态。
	return r.width * r.height
}

func (r rect) perim() int {
	// 值接收者适合只读且对象较小的场景。
	return 2*r.width + 2*r.height
}

func TestStructMethod(t *testing.T) {
	r := rect{width: 10, height: 5}

	area := r.area()
	fmt.Println(area)

	perim := r.perim()
	fmt.Println(perim)

	// 延伸练习：
	// 1. 给 rect 再加一个 scale 方法，尝试修改长宽。
	// 2. 比较 scale 用值接收者和指针接收者时的行为差异。
}
