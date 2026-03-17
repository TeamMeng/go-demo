package main

import "testing"

func TestGreet(t *testing.T) {
	got := greet("GitHub")
	want := "Hello, GitHub! GitHub Actions is running."

	if got != want {
		t.Fatalf("greet() = %q, want %q", got, want)
	}
}
