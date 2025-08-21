package main

func main() {
	r := condSelect(123, 456)
	println("result is %d", r)
}

func condSelect(a, b int) int {
	var c int
	if a > b {
		c = a + 2
	} else {
		c = b + 3
	}
	return c
}
