package main

func main() {
	for i := 0; i < 10; i++ {
		r := condSelect(i%5, i%3)
		println("result is %d", r)
	}
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
