package experiments

import "testing"

func vals() (int, int) {
	return 3, 7
}

func TestMultipleReturnValues(t *testing.T) {
	// Go 可以一次返回多个值。
	a, b := vals()
	expectedA, expectedB := 3, 7
	if a != expectedA || b != expectedB {
		t.Errorf("Expected %d and %d, but got %d and %d", expectedA, expectedB, a, b)
	}

	// 延伸练习：
	// 1. 把其中一个返回值改成 error，模拟更常见的 Go 风格。
	// 2. 使用 _ 丢弃一个不需要的返回值。
}
