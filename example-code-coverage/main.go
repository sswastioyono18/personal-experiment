package example_code_coverage

import "fmt"

func main() {
	fmt.Println(Sum(1, 2))
}

func Sum(a, b int) int {
	return a + b
}
