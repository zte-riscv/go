// asmcheck

package codegen

func Zicond(a, b int) int {
	var c int
	if a > b {
		c = a
	} else {
		c = b
	}
	// riscv64:'CZEROE2Z'
	return c
}
