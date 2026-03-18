package experiments

import (
	"fmt"
	"testing"
)

func TestMap(t *testing.T) {
	// map 必须先初始化后才能写入。
	m := make(map[string]int)
	fmt.Println(m)

	m["one"] = 1
	m["two"] = 2
	fmt.Println(m)

	value := m["one"]
	fmt.Println(value)

	delete(m, "two")
	fmt.Println(m)

	// 不存在的 key 会返回 value 类型的零值。
	value = m["two"]
	fmt.Println(value)

	// clear 会清空整个 map。
	clear(m)
	fmt.Println(m)

	// 延伸练习：
	// 1. 使用 value, ok := m[key] 判断 key 是否存在。
	// 2. 尝试遍历 map，观察输出顺序是否固定。
}
