package experiments

import "testing"

func printHelloWorld(input string) string {
	return "Hello, " + input + "!"
}

func TestFunc(t *testing.T) {
	// 调用普通函数并校验返回值。
	result := printHelloWorld("World")
	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}

	// 延伸练习：
	// 1. 把函数改成接收 firstName 和 lastName 两个参数。
	// 2. 为不同输入补两组三表驱动测试。
}
