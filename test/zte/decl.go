//go:build riscv64

package zte

//go:noescape
func vlseg2E8VAndVsseg2E8VForDeinterleaveAndInterleave(in []byte, out0 []byte, out1 []byte, out3 []byte)
