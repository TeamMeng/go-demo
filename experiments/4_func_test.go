package experiments

import "testing"

func printHelloWorld(input string) string {
	return "Hello, " + input + "!"
}

func TestFunc(t *testing.T) {
	result := printHelloWorld("World")
	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}
