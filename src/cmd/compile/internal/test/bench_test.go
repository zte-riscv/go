// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import "testing"

var globl int64
var globl32 int32

func BenchmarkLoadAdd(b *testing.B) {
	x := make([]int64, 1024)
	y := make([]int64, 1024)
	for i := 0; i < b.N; i++ {
		var s int64
		for i := range x {
			s ^= x[i] + y[i]
		}
		globl = s
	}
}

// Added for ppc64 extswsli on power9
func BenchmarkExtShift(b *testing.B) {
	x := make([]int32, 1024)
	for i := 0; i < b.N; i++ {
		var s int64
		for i := range x {
			s ^= int64(x[i]+32) * 8
		}
		globl = s
	}
}

func BenchmarkModify(b *testing.B) {
	a := make([]int64, 1024)
	v := globl
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] += v
		}
	}
}

func BenchmarkMullImm(b *testing.B) {
	x := make([]int32, 1024)
	for i := 0; i < b.N; i++ {
		var s int32
		for i := range x {
			s += x[i] * 100
		}
		globl32 = s
	}
}

func BenchmarkConstModify(b *testing.B) {
	a := make([]int64, 1024)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] += 3
		}
	}
}

func BenchmarkBitSet(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < N; j++ {
			a[j/64] |= 1 << (j % 64)
		}
	}
}

func BenchmarkBitClear(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < N; j++ {
			a[j/64] &^= 1 << (j % 64)
		}
	}
}

func BenchmarkBitToggle(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < N; j++ {
			a[j/64] ^= 1 << (j % 64)
		}
	}
}

func BenchmarkBitSetConst(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] |= 1 << 37
		}
	}
}

func BenchmarkBitClearConst(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] &^= 1 << 37
		}
	}
}

func BenchmarkBitToggleConst(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] ^= 1 << 37
		}
	}
}

func BenchmarkMul64Const3(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 3
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const5(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 5
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const6(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 6
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const8(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 8
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const9(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 9
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const10(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 10
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const12(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 12
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul32Const3(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 3
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const6(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 6
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const10(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 10
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul64Const18(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 18
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const20(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 20
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const24(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 24
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const36(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 36
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const40(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 40
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64Const72(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 72
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

// 64-bit negative SH*ADD patterns.

func BenchmarkMul64ConstNeg3(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -3
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64ConstNeg5(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -5
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64ConstNeg6(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -6
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64ConstNeg9(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -9
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64ConstNeg10(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -10
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64ConstNeg12(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -12
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul64ConstNeg18(b *testing.B) {
	var s, x int64
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -18
		x = (s ^ int64(i)) & 0xfff
	}
	globl = s
}

func BenchmarkMul32Const5(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 5
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const9(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 9
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const12(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 12
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const18(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 18
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const20(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 20
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32Const36(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * 36
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32ConstNeg3(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -3
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32ConstNeg5(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -5
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}

func BenchmarkMul32ConstNeg9(b *testing.B) {
	var s, x int32
	x = 7
	for i := 0; i < b.N; i++ {
		s += x * -9
		x = int32((int64(s) ^ int64(i)) & 0xfff)
	}
	globl32 = s
}
