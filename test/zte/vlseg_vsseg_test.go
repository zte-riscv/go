//go:build riscv64

package zte

import (
	"slices"
	"testing"
)

func TestVLSEG2E8AndVSSEG2E8VByDeinterleaveAndInterleave(t *testing.T) {
	const pairs = 10
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

	out3 := make([]byte, pairs*2)

	vlseg2E8VAndVsseg2E8VForDeinterleaveAndInterleave(in, out0, out1, out3)

	if string(out0) != string(exp0) {
		t.Fatalf("segment0 mismatch, in: %v, out0: %v, exp0: %v", in, out0, exp0)
	}
	if string(out1) != string(exp1) {
		t.Fatalf("segment1 mismatch, in: %v, out1: %v, exp1: %v", in, out1, exp1)
	}
	if string(out3) != string(in) {
		t.Fatalf("out3 mismatch, in: %v, out3: %v", in, out3)
	}
	t.Logf("in: %v, out3: %v, out0: %v, exp0: %v, out1: %v, exp1: %v", in, out3, out0, exp0, out1, exp1)
}

func TestVLSEG3E8AndVSSEG3E8VByDeinterleaveAndInterleave(t *testing.T) {
	const pairs = 10
	in := make([]byte, pairs*3)
	exp0 := make([]byte, pairs)
	exp1 := make([]byte, pairs)
	exp2 := make([]byte, pairs)
	for i := 0; i < pairs; i++ {
		a := byte(i)
		b := byte(200 - i)
		c := byte(300 - i)
		in[3*i+0] = a
		in[3*i+1] = b
		in[3*i+2] = c
		exp0[i] = a
		exp1[i] = b
		exp2[i] = c
	}

	out0 := make([]byte, pairs)
	out1 := make([]byte, pairs)
	out2 := make([]byte, pairs)

	out3 := make([]byte, pairs*3)

	vlseg3E8VAndVsseg3E8VForDeinterleaveAndInterleave(in, out0, out1, out2, out3)

	if string(out0) != string(exp0) {
		t.Fatalf("segment0 mismatch, in: %v, out0: %v, exp0: %v", in, out0, exp0)
	}
	if string(out1) != string(exp1) {
		t.Fatalf("segment1 mismatch, in: %v, out1: %v, exp1: %v", in, out1, exp1)
	}
	if string(out2) != string(exp2) {
		t.Fatalf("segment2 mismatch, in: %v, out2: %v, exp2: %v", in, out2, exp2)
	}

	if string(out3) != string(in) {
		t.Fatalf("out3 mismatch, in: %v, out3: %v", in, out3)
	}
	t.Logf("in: %v, out3: %v, out0: %v, exp0: %v, out1: %v, exp1: %v, out2: %v, exp2: %v", in, out3, out0, exp0, out1, exp1, out2, exp2)
}

func TestVLSEG2E32AndVSSEG2E32VByDeinterleaveAndInterleave(t *testing.T) {
	const pairs = 10
	in := make([]int32, pairs*2)
	exp0 := make([]int32, pairs)
	exp1 := make([]int32, pairs)
	for i := 0; i < pairs; i++ {
		a := int32(i)
		b := int32(200 - i)
		in[2*i+0] = a
		in[2*i+1] = b
		exp0[i] = a
		exp1[i] = b
	}

	out0 := make([]int32, pairs)
	out1 := make([]int32, pairs)

	out3 := make([]int32, pairs*2)

	vlseg2E32VAndVsseg2E32VForDeinterleaveAndInterleave(in, out0, out1, out3)

	if !slices.Equal(out0, exp0) {
		t.Fatalf("segment0 mismatch, in: %v, out0: %v, exp0: %v", in, out0, exp0)
	}
	if !slices.Equal(out1, exp1) {
		t.Fatalf("segment1 mismatch, in: %v, out1: %v, exp1: %v", in, out1, exp1)
	}
	if !slices.Equal(out3, in) {
		t.Fatalf("out3 mismatch, in: %v, out3: %v", in, out3)
	}
	t.Logf("in: %v, out3: %v, out0: %v, exp0: %v, out1: %v, exp1: %v", in, out3, out0, exp0, out1, exp1)
}

func TestVLSEG2E64AndVSSEG2E64VByDeinterleaveAndInterleave(t *testing.T) {
	const pairs = 10
	in := make([]int64, pairs*2)
	exp0 := make([]int64, pairs)
	exp1 := make([]int64, pairs)
	for i := 0; i < pairs; i++ {
		a := int64(i)
		b := int64(200 - i)
		in[2*i+0] = a
		in[2*i+1] = b
		exp0[i] = a
		exp1[i] = b
	}

	out0 := make([]int64, pairs)
	out1 := make([]int64, pairs)

	out3 := make([]int64, pairs*2)

	vlseg2E64VAndVsseg2E64VForDeinterleaveAndInterleave(in, out0, out1, out3)

	if !slices.Equal(out0, exp0) {
		t.Fatalf("segment0 mismatch, in: %v, out0: %v, exp0: %v", in, out0, exp0)
	}
	if !slices.Equal(out1, exp1) {
		t.Fatalf("segment1 mismatch, in: %v, out1: %v, exp1: %v", in, out1, exp1)
	}
	if !slices.Equal(out3, in) {
		t.Fatalf("out3 mismatch, in: %v, out3: %v", in, out3)
	}
	t.Logf("in: %v, out3: %v, out0: %v, exp0: %v, out1: %v, exp1: %v", in, out3, out0, exp0, out1, exp1)
}
