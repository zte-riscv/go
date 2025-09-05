//go:build riscv64

package zte

//go:noescape
func vlseg2Deinterleave(in []byte, out0 []byte, out1 []byte)
