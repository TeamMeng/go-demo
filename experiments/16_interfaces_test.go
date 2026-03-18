package experiments

import (
	"fmt"
	"math"
	"testing"
)

type geometry interface {
	area() float64
	perimeter() float64
}

type rectangle struct {
	width  float64
	height float64
}

type circle struct {
	radius float64
}

func (r rectangle) area() float64 {
	return r.width * r.height
}

func (r rectangle) perimeter() float64 {
	return 2 * (r.width + r.height)
}

func (c circle) area() float64 {
	return math.Pi * c.radius * c.radius
}

func (c circle) perimeter() float64 {
	return 2 * math.Pi * c.radius
}

func TestInterface(t *testing.T) {
	// 两个不同类型，只要方法集一致，就能表达相同抽象能力。
	r := rectangle{width: 3.0, height: 4.0}
	c := circle{radius: 5.0}

	fmt.Printf("rectangle area %f and perimeter %f\n", r.area(), r.perimeter())
	fmt.Printf("circle area %f and perimeter %f\n", c.area(), c.perimeter())

	// 延伸练习：
	// 1. 写一个 measure(g geometry) 函数，真正以接口作为参数。
	// 2. 再增加一个 triangle 类型，让它也实现 geometry。
}
