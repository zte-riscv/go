//go:build riscv64

package zte

import (
	"testing"
)

func TestVLSEG2E8_Deinterleave(t *testing.T) {

	// 构造交错输入: [a0,b0,a1,b1,...]
	const pairs = 64
	in := make([]byte, pairs*2)
	exp0 := make([]byte, pairs)
	exp1 := make([]byte, pairs)
	for i := 0; i < pairs; i++ {
		a := byte(i)
		b := byte(200 - i)
		in[2*i+0] = a
		in[2*i+1] = b
		exp0[i] = a
		exp1[i] = b
	}

	out0 := make([]byte, pairs)
	out1 := make([]byte, pairs)

	vlseg2Deinterleave(in, out0, out1)

	if string(out0) != string(exp0) {
		t.Fatalf("segment0 mismatch, in: %v, out0: %v, exp0: %v, out1: %v, exp1: %v", in, out0, exp0, out1, exp1)
	}
	if string(out1) != string(exp1) {
		t.Fatalf("segment1 mismatch, in: %v, out0: %v, exp0: %v, out1: %v, exp1: %v", in, out0, exp0, out1, exp1)
	}
}
