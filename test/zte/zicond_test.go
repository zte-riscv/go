package zte

import "testing"

//go:noinline
func Zicond(a, b int) int {
	var c int
	if a > b {
		c = a
	} else {
		c = b
	}
	return c
}
func TestZicond(t *testing.T) {
	for i := 0; i < 1000000; i++ {
		rst := Zicond(i, i+1)
		if rst != i+1 {
			t.Errorf("Zicond(%d, %d) = %d, want %d", i, i+1, rst, i+1)
		}
	}
}
