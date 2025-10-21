//go:build riscv64

package asmtest

import (
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	myf()
	fmt.Println(1)
}
