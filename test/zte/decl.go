//go:build riscv64

package zte

//go:noescape
func vlseg2E8VAndVsseg2E8VForDeinterleaveAndInterleave(in []byte, out0 []byte, out1 []byte, out3 []byte)

func vlseg3E8VAndVsseg3E8VForDeinterleaveAndInterleave(in []byte, out0 []byte, out1 []byte, out2 []byte, out3 []byte)

func vlseg2E32VAndVsseg2E32VForDeinterleaveAndInterleave(in []int32, out0 []int32, out1 []int32, out3 []int32)

func vlseg2E64VAndVsseg2E64VForDeinterleaveAndInterleave(in []int64, out0 []int64, out1 []int64, out3 []int64)
