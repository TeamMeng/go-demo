package main

import "testing"

func TestGreet(t *testing.T) {
	// 测试直接覆盖 greet 的对外行为，避免 main 中的打印逻辑干扰断言。
	got := greet("GitHub")
	want := "Hello, GitHub! GitHub Actions is running."

	if got != want {
		t.Fatalf("greet() = %q, want %q", got, want)
	}
}
