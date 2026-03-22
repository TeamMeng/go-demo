package main

import "fmt"

// greet 生成一个固定格式的问候语，方便在 main 和测试里复用。
func greet(name string) string {
	return fmt.Sprintf("Hello, %s! GitHub Actions is running.", name)
}

func main() {
	// 主程序只负责调用核心逻辑并输出结果。
	fmt.Println(greet("Go"))
}
