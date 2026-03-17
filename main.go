package main

import "fmt"

func greet(name string) string {
	return fmt.Sprintf("Hello, %s! GitHub Actions is running.", name)
}

func main() {
	fmt.Println(greet("Go"))
}
